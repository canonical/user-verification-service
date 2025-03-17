package directoryapi

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client struct {
	c *http.Client

	url string
}

func (c *Client) IsEmployee(ctx context.Context, mail string) (bool, error) {
	return false, nil
}

func NewClient(url string) *Client {
	c := new(Client)

	c.c = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	c.url = url
	return c
}
