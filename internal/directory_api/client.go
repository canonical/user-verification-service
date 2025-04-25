package directoryapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/canonical/user-verification-service/internal/monitoring"
	"github.com/canonical/user-verification-service/internal/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var ErrInvalidApiToken = fmt.Errorf("invalid api token")
var ErrUnknownApiError = fmt.Errorf("unknown api error")

type Client struct {
	http HttpClientInterface

	url   string
	token string

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (c *Client) IsEmployee(ctx context.Context, mail string) (bool, error) {
	r, err := http.NewRequestWithContext(ctx, "GET", c.url, nil)
	if err != nil {
		c.logger.Errorf("Failed to construct request: %d", err)
		return false, err
	}

	r.Header.Add("Authorization", fmt.Sprint("Bearer ", c.token))
	q := r.URL.Query()
	q.Add("email", mail)
	r.URL.RawQuery = q.Encode()

	rr, err := c.http.Do(r)
	if err != nil {
		return false, err
	}

	switch rr.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	case http.StatusUnauthorized:
		c.logger.Errorf("Got status code %s, the API token is invalid", rr.StatusCode)
		return false, ErrInvalidApiToken
	default:
		c.logger.Errorf("Got unexpected status code: %d", rr.StatusCode)
		return false, ErrUnknownApiError
	}
}

func NewClient(
	skipTlsVerification bool,
	u, token string,
	tracer tracing.TracingInterface,
	monitor monitoring.MonitorInterface,
	logger logging.LoggerInterface,
) *Client {
	c := new(Client)

	cc := http.DefaultTransport.(*http.Transport).Clone()
	cc.TLSClientConfig = &tls.Config{InsecureSkipVerify: skipTlsVerification}
	c.http = &http.Client{Transport: otelhttp.NewTransport(cc)}

	c.token = token
	c.url = u

	c.monitor = monitor
	c.tracer = tracer
	c.logger = logger

	return c
}
