package services

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/alitto/pond/v2"
	"github.com/duke-git/lancet/v2/condition"
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"github.com/muety/wakapi/utils/cache"
)

const (
	gridRows      = 7
	cellWidth     = 20
	cellHeight    = 20
	cellSpacing   = 3
	colorMinDark  = "#242B3A"
	colorMinLight = "#DCE3E1"
	colorMaxDark  = "#047857"
	colorMaxLight = "#047857"
	textDark      = "#D1D5DB"
	textLight     = "#37474F"
)

type ActivityService struct {
	config         *config.Config
	cache          *cache.Cache
	summaryService ISummaryService
}

func NewActivityService(summaryService ISummaryService) *ActivityService {
	return &ActivityService{
		config:         config.Get(),
		cache:          cache.New(6*time.Hour, 6*time.Hour),
		summaryService: summaryService,
	}
}

// GetChart generates an activity chart for a given user and the given time interval, similar to GitHub's contribution timeline. See https://github.com/muety/wakapi/issues/12.
// Please note: currently, only yearly charts ("last_12_months") are supported. However, we could fairly easily restructure this to support dynamic intervals.
func (s *ActivityService) GetChart(user *models.User, interval *models.IntervalKey, darkTheme, hideAttribution, skipCache bool) (string, error) {
	cacheKey := fmt.Sprintf("chart_%s_%s_%v_%v", user.ID, (*interval)[0], darkTheme, hideAttribution)
	if result, found := s.cache.Get(cacheKey); found && !skipCache {
		return result.(string), nil
	}

	switch interval {
	case models.IntervalPast12Months:
		chart, err := s.getChartPastYear(user, darkTheme, hideAttribution)
		if err == nil {
			s.cache.SetDefault(cacheKey, chart) // TODO: cache compressed?
		}
		return chart, err
	default:
		return "", errors.New("unsupported interval")
	}
}

func (s *ActivityService) getChartPastYear(user *models.User, darkTheme, hideAttribution bool) (string, error) {
	err, from, to := helpers.ResolveIntervalTZ(models.IntervalPast12Months, user.TZ(), user.StartOfWeekDay())
	// TODO: I am not sure if we have to handle startOfWeekDay here, but it seems like we do not need to.
	from = datetime.BeginOfWeek(from, time.Monday)
	if err != nil {
		return "", err
	}

	intervals := utils.SplitRangeByDays(from, to)
	summaries := make([]*models.Summary, len(intervals))

	wp := pond.NewPool(utils.HalfCPUs())
	mut := sync.RWMutex{}

	// fetch summaries
	for i, interval := range intervals {
		i := i // https://github.com/golang/go/wiki/CommonMistakes#using-reference-to-loop-iterator-variable
		interval := interval

		wp.Submit(func() {
			summary, err := s.summaryService.Retrieve(interval[0], interval[1], user, nil, nil)
			if err != nil {
				config.Log().Warn("failed to retrieve summary for activity chart", "userID", user.ID, "from", from, "to", to)
				summary = models.NewEmptySummary()
				summary.FromTime = models.CustomTime(from)
				summary.ToTime = models.CustomTime(to)
				summary.UserID = user.ID
				summary.User = user
			}

			mut.Lock()
			summaries[i] = summary
			mut.Unlock()
		})
	}

	wp.StopAndWait()

	maxTotal := models.Summaries(summaries).MaxTotalTime()

	var (
		colorRGBAMin         = utils.HexToRGBA(condition.Ternary[bool, string](darkTheme, colorMinDark, colorMinLight))
		colorRGBAMax         = utils.HexToRGBA(condition.Ternary[bool, string](darkTheme, colorMaxDark, colorMaxLight))
		colorText            = condition.Ternary[bool, string](darkTheme, textDark, textLight)
		gridCols             = math.Ceil(float64(len(summaries)) / float64(gridRows))
		w            float64 = gridCols*cellWidth + gridCols*cellSpacing
		h            float64 = gridRows*cellHeight + 25 + 24 + 5 + 5 + gridRows*cellSpacing
	)

	// regenerate svg
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f">`, w, h)
	fmt.Fprintf(
		buf,
		`<style type="text/css">text { font-family: 'Source Sans 3', Roboto, Helvetica, Arial, sans-serif; font-size: 0.9rem; font-weight: 500; fill: %s; } rect { fill-opacity: 1; rx: 3px; ry: 3px; } rect:hover { filter: brightness(0.9) }</style>`,
		colorText,
	)
	fmt.Fprintf(
		buf,
		`<text x="0" y="15">%s</text>`,
		escapeSVGText(fmt.Sprintf("%s to %s", helpers.FormatDateHuman(summaries[0].FromTime.T()), helpers.FormatDateHuman(summaries[len(summaries)-1].ToTime.T()))),
	)

	for i, s := range summaries {
		total := s.TotalTime()
		fillColor := utils.RGBAToHex(utils.FadeColors(colorRGBAMin, colorRGBAMax, float64(total)/float64(maxTotal)))
		title := fmt.Sprintf("%s on %s", helpers.FmtWakatimeDuration(total), helpers.FormatDateHuman(s.FromTime.T()))
		x := float64(i/gridRows) * (cellWidth + cellSpacing)
		y := 25 + float64((i%gridRows)*(cellHeight+cellSpacing))
		fmt.Fprintf(
			buf,
			`<g><title>%s</title><rect x="%.0f" y="%.0f" width="%d" height="%d" style="fill: %s" /></g>`,
			escapeSVGText(title),
			x,
			y,
			cellWidth,
			cellHeight,
			fillColor,
		)
	}

	if !hideAttribution {
		fmt.Fprintf(
			buf,
			`<g><title>Wakapi.dev</title><image x="%.0f" y="%.0f" width="60" height="24" href="https://wakapi.dev/assets/images/logo-gh.svg" /></g>`,
			w-60,
			h-24,
		)
	}

	buf.WriteString(`</svg>`)

	return buf.String(), nil
}

func escapeSVGText(value string) string {
	var escaped bytes.Buffer
	_ = xml.EscapeText(&escaped, []byte(value))
	return escaped.String()
}
