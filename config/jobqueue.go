package config

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alitto/pond/v2"
	"github.com/muety/wakapi/utils"
	"github.com/robfig/cron/v3"
)

var jobQueues map[string]*JobQueue
var jobCounts map[string]int

const (
	QueueDefault      = "wakapi.default"
	QueueProcessing   = "wakapi.processing"
	QueueProcessing2  = "wakapi.processing_2"
	QueueReports      = "wakapi.reports"
	QueueMails        = "wakapi.mail"
	QueueImports      = "wakapi.imports"
	QueueHousekeeping = "wakapi.housekeeping"
)

type JobQueueMetrics struct {
	Queue        string
	EnqueuedJobs int
	FinishedJobs int
}

func init() {
	jobQueues = make(map[string]*JobQueue)
}

type JobQueue struct {
	name       string
	pool       pond.Pool
	stop       chan struct{}
	stopOnce   sync.Once
	enqueued   atomic.Int32
	dispatched atomic.Int32
}

type DispatchTicker struct {
	stop     chan struct{}
	stopOnce sync.Once
}

type DispatchCron struct {
	cron *cron.Cron
}

func StartJobs() {
	InitQueue(QueueDefault, 1)
	InitQueue(QueueProcessing, utils.HalfCPUs())
	InitQueue(QueueProcessing2, utils.HalfCPUs())
	InitQueue(QueueReports, 1)
	InitQueue(QueueMails, 1)
	InitQueue(QueueImports, 1)
	InitQueue(QueueHousekeeping, utils.HalfCPUs())
}

func InitQueue(name string, workers int) error {
	if _, ok := jobQueues[name]; ok {
		return fmt.Errorf("queue '%s' already existing", name)
	}
	slog.Info("creating job queue", "name", name, "workers", workers)
	jobQueues[name] = NewJobQueue(name, workers, 4096)
	return nil
}

func NewJobQueue(name string, workers, queueSize int) *JobQueue {
	return &JobQueue{
		name: name,
		pool: pond.NewPool(
			workers,
			pond.WithQueueSize(queueSize),
		),
		stop: make(chan struct{}),
	}
}

func (q *JobQueue) Dispatch(run func()) error {
	q.enqueued.Add(1)
	err := q.pool.Go(func() {
		defer q.enqueued.Add(-1)
		defer q.dispatched.Add(1)
		run()
	})
	if err != nil {
		q.enqueued.Add(-1)
	}
	return err
}

func (q *JobQueue) DispatchIn(run func(), duration time.Duration) error {
	timer := time.NewTimer(duration)
	go func() {
		select {
		case <-timer.C:
			if err := q.Dispatch(run); err != nil {
				slog.Error("failed to dispatch delayed job", "queue", q.name, "error", err)
			}
		case <-q.stop:
			timer.Stop()
		}
	}()
	return nil
}

func (q *JobQueue) DispatchEvery(run func(), interval time.Duration) (*DispatchTicker, error) {
	ticker := time.NewTicker(interval)
	handle := &DispatchTicker{stop: make(chan struct{})}
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := q.Dispatch(run); err != nil {
					slog.Error("failed to dispatch periodic job", "queue", q.name, "error", err)
				}
			case <-handle.stop:
				return
			case <-q.stop:
				return
			}
		}
	}()
	return handle, nil
}

func (q *JobQueue) DispatchCron(run func(), cronStr string) (*DispatchCron, error) {
	c := cron.New()
	if _, err := c.AddFunc(cronStr, func() {
		if err := q.Dispatch(run); err != nil {
			slog.Error("failed to dispatch cron job", "queue", q.name, "error", err)
		}
	}); err != nil {
		return nil, err
	}
	c.Start()
	return &DispatchCron{cron: c}, nil
}

func (q *JobQueue) CountEnqueued() int {
	return int(q.enqueued.Load())
}

func (q *JobQueue) CountDispatched() int {
	return int(q.dispatched.Load())
}

func (q *JobQueue) Stop() {
	q.stopOnce.Do(func() {
		close(q.stop)
		q.pool.StopAndWait()
	})
}

func (t *DispatchTicker) Stop() {
	t.stopOnce.Do(func() {
		close(t.stop)
	})
}

func (c *DispatchCron) Stop() {
	if c == nil || c.cron == nil {
		return
	}
	<-c.cron.Stop().Done()
}

func GetDefaultQueue() *JobQueue {
	return GetQueue(QueueDefault)
}

func GetQueue(name string) *JobQueue {
	if _, ok := jobQueues[name]; !ok {
		InitQueue(name, 1)
	}
	return jobQueues[name]
}

func GetQueueMetrics() []*JobQueueMetrics {
	metrics := make([]*JobQueueMetrics, 0, len(jobQueues))
	for name, queue := range jobQueues {
		metrics = append(metrics, &JobQueueMetrics{
			Queue:        name,
			EnqueuedJobs: queue.CountEnqueued(),
			FinishedJobs: queue.CountDispatched(),
		})
	}
	return metrics
}

func CloseQueues() {
	for _, q := range jobQueues {
		q.Stop()
	}
}
