package directoryapi

import (
	"context"
	"net/http"
)

type DirectoryAPI interface {
	IsEmployee(context.Context, string) (bool, error)
}

type HttpClientInterface interface {
	Do(*http.Request) (*http.Response, error)
}
