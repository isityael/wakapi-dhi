package mail

/*
Uses smtp4Dev (https://github.com/rnwood/smtp4dev) mock SMTP server to test against.
To spawn an smtp4Dev instance in Docker, run:
$ docker run --rm -it -p 5000:80 -p 2525:25 -p 8080:80 rnwood/smtp4Dev
*/

// TODO: test actual message content / title / recipients / etc.
// TODO: run in ci (gh actions)

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log/slog"
)

const (
	TestSmtpUser         = "admin"
	TestSmtpPass         = "admin"
	Smtp4DevApiUrl       = "http://localhost:8080/api"
	Smtp4DevHost         = "localhost"
	Smtp4DevPort         = 2525
	smtpReadinessRetries = 30
	smtpReadinessDelay   = time.Second
)

type SmtpTestSuite struct {
	suite.Suite
	smtp4Dev *Smtp4DevClient
}

func (suite *SmtpTestSuite) SetupSuite() {
	suite.smtp4Dev = newSmtp4DevClient()
	if err := suite.smtp4Dev.Setup(); err != nil {
		suite.Error(err)
	}
}

func (suite *SmtpTestSuite) BeforeTest(suiteName, testName string) {
	if err := suite.smtp4Dev.ClearInboxes(); err != nil {
		suite.Error(err)
	}
}

func TestSmtpTestSuite(t *testing.T) {
	address := net.JoinHostPort(smtp4DevHost(), fmt.Sprintf("%d", smtp4DevPort()))
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Skipf("WARNING: smtp4Dev not available at %s - skipping smtp tests", address)
		return
	}
	conn.Close()

	smtp4dev := newSmtp4DevClient()
	for i := 0; i < 5; i++ {
		if smtp4dev.Check() == nil {
			break
		}
		t.Logf("smtp4Dev not ready at %s, retrying...", smtp4dev.ApiBaseUrl)
		time.Sleep(1 * time.Second)
	}

	suite.Run(t, new(SmtpTestSuite))
}

func TestRetrySmtpReadiness(t *testing.T) {
	attempts := 0
	err := retrySmtpReadiness(3, 0, func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("SMTP listener is still restarting")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestRetrySmtpReadinessTimesOut(t *testing.T) {
	attempts := 0
	err := retrySmtpReadiness(2, 0, func() error {
		attempts++
		return fmt.Errorf("SMTP protocol not ready")
	})

	assert.EqualError(t, err, "SMTP protocol not ready after 2 attempts: SMTP protocol not ready")
	assert.Equal(t, 2, attempts)
}

func (suite *SmtpTestSuite) TestSMTPSendingService_SendPlain() {
	if err := suite.smtp4Dev.SetNoTls(); err != nil {
		suite.Error(err)
	}

	cfg := createDefaultSMTPConfig()

	sut := NewSMTPSendingService(cfg)
	err := sut.Send(createTestMail())

	msgCount, _ := suite.smtp4Dev.CountMessages()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, msgCount)
}

func (suite *SmtpTestSuite) TestSMTPSendingService_SendTLS() {
	if err := suite.smtp4Dev.SetForcedTls(); err != nil {
		suite.Error(err)
	}

	cfg := createDefaultSMTPConfig()
	cfg.TLS = true

	sut := NewSMTPSendingService(cfg)
	err := sut.Send(createTestMail())

	msgCount, _ := suite.smtp4Dev.CountMessages()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, msgCount)
}

func (suite *SmtpTestSuite) TestSMTPSendingService_SendStartTLS() {
	if err := suite.smtp4Dev.SetStartTls(); err != nil {
		suite.Error(err)
	}

	cfg := createDefaultSMTPConfig()
	cfg.TLS = false

	sut := NewSMTPSendingService(cfg)
	err := sut.Send(createTestMail())

	msgCount, _ := suite.smtp4Dev.CountMessages()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, msgCount)
}

// Private utility methods

func createTestMail() *models.Mail {
	mail := models.Mail{
		From:    "Wakapi <noreply@wakapi.dev>",
		To:      []models.MailAddress{"Ferdinand Mütsch <ferdinand@muetsch.io>"},
		Subject: "Wakapi Test Mail",
		Body:    "This is just a test",
		Type:    models.PlainType,
		Date:    time.Now(),
	}
	return mail.Sanitized()
}

func createDefaultSMTPConfig() config.SMTPMailConfig {
	return config.SMTPMailConfig{
		Host:       smtp4DevHost(),
		Port:       smtp4DevPort(),
		Username:   TestSmtpUser,
		Password:   TestSmtpPass,
		TLS:        false,
		SkipVerify: true,
	}
}

type Smtp4DevClient struct {
	ApiBaseUrl string
}

func newSmtp4DevClient() *Smtp4DevClient {
	return &Smtp4DevClient{
		ApiBaseUrl: smtp4DevApiUrl(),
	}
}

func smtp4DevApiUrl() string {
	if value := os.Getenv("WAKAPI_TEST_SMTP4DEV_API_URL"); value != "" {
		return value
	}
	return Smtp4DevApiUrl
}

func smtp4DevHost() string {
	if value := os.Getenv("WAKAPI_TEST_SMTP4DEV_HOST"); value != "" {
		return value
	}
	return Smtp4DevHost
}

func smtp4DevPort() uint {
	if value := os.Getenv("WAKAPI_TEST_SMTP4DEV_PORT"); value != "" {
		port, err := strconv.ParseUint(value, 10, 16)
		if err == nil {
			return uint(port)
		}
	}
	return Smtp4DevPort
}

func (c *Smtp4DevClient) Check() error {
	res, err := http.Get(fmt.Sprintf("%s/Version", c.ApiBaseUrl))
	if err != nil {
		return err
	}
	if _, err := utils.RaiseForStatus(res, err); err != nil {
		return err
	}
	return nil
}

func (c *Smtp4DevClient) Setup() error {
	if c.Check() != nil {
		return fmt.Errorf("smtp4Dev is unavailable at %s", c.ApiBaseUrl)
	}

	if err := c.SetConfigValue("deliverMessagesToUsersDefaultMailbox", false); err != nil {
		return err
	}
	if err := c.waitForSmtpMode(false); err != nil {
		return err
	}

	if err := c.CreateTestUsers(); err != nil {
		return err
	}

	if err := c.ClearInboxes(); err != nil {
		return err
	}

	return nil
}

func (c *Smtp4DevClient) GetConfig() (map[string]interface{}, error) {
	var data map[string]interface{}

	res, err := http.Get(fmt.Sprintf("%s/Server", c.ApiBaseUrl))
	if err != nil {
		return nil, err
	}
	if _, err := utils.RaiseForStatus(res, err); err != nil {
		return nil, err
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Smtp4DevClient) CountMessages() (int, error) {
	var data map[string]interface{}

	res, err := http.Get(fmt.Sprintf("%s/Messages", c.ApiBaseUrl))
	if err != nil {
		return 0, err
	}
	if _, err := utils.RaiseForStatus(res, err); err != nil {
		return 0, err
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return 0, err
	}

	return len(data["results"].([]interface{})), nil
}

func (c *Smtp4DevClient) SetNoTls() error {
	slog.Info("[smtp4Dev] disabling tls encryption")
	if err := c.SetConfigValue("tlsMode", "None"); err != nil {
		return err
	}
	return c.waitForSmtpMode(false)
}

func (c *Smtp4DevClient) SetForcedTls() error {
	slog.Info("[smtp4Dev] enabling forced tls encryption")
	if err := c.SetConfigValue("tlsMode", "ImplicitTls"); err != nil {
		return err
	}
	return c.waitForSmtpMode(true)
}

func (c *Smtp4DevClient) SetStartTls() error {
	slog.Info("[smtp4Dev] enabling tls encryption via starttls")
	if err := c.SetConfigValue("tlsMode", "StartTls"); err != nil {
		return err
	}
	return c.waitForSmtpMode(false)
}

func (c *Smtp4DevClient) CreateTestUsers() error {
	slog.Info("[smtp4Dev] creating test users")
	if err := c.SetConfigValue("users", []map[string]interface{}{
		{
			"username":       TestSmtpUser,
			"password":       TestSmtpPass,
			"defaultMailbox": "Default",
		},
	}); err != nil {
		return err
	}
	return c.waitForSmtpMode(false)
}

func (c *Smtp4DevClient) ClearInboxes() error {
	slog.Info("[smtp4Dev] clearing inboxes")
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/Messages/*", c.ApiBaseUrl), nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if _, err := utils.RaiseForStatus(res, err); err != nil {
		return err
	}

	return nil
}

func (c *Smtp4DevClient) SetConfigValue(key string, val interface{}) error {
	settings, err := c.GetConfig()
	if err != nil {
		return err
	}

	settings[key] = val

	data, _ := json.Marshal(settings)
	res, err := http.Post(fmt.Sprintf("%s/Server", c.ApiBaseUrl), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	if _, err := utils.RaiseForStatus(res, err); err != nil {
		return err
	}

	return nil
}

func (c *Smtp4DevClient) waitForSmtpMode(implicitTLS bool) error {
	address := net.JoinHostPort(smtp4DevHost(), fmt.Sprintf("%d", smtp4DevPort()))
	return retrySmtpReadiness(smtpReadinessRetries, smtpReadinessDelay, func() error {
		dialer := &net.Dialer{Timeout: time.Second}
		if implicitTLS {
			conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{ //nolint:gosec // smtp4dev uses a self-signed test certificate
				InsecureSkipVerify: true,
			})
			if err != nil {
				return err
			}
			return conn.Close()
		}

		conn, err := dialer.Dial("tcp", address)
		if err != nil {
			return err
		}
		defer conn.Close()
		if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			return err
		}
		banner := make([]byte, 4)
		if _, err := io.ReadFull(conn, banner); err != nil {
			return err
		}
		if !bytes.Equal(banner, []byte("220 ")) {
			return fmt.Errorf("unexpected SMTP banner %q", banner)
		}
		return nil
	})
}

func retrySmtpReadiness(attempts int, delay time.Duration, probe func() error) error {
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if lastErr = probe(); lastErr == nil {
			return nil
		}
		if attempt < attempts {
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("SMTP protocol not ready after %d attempts: %w", attempts, lastErr)
}
