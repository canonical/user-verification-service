package config

import "flag"

// EnvSpec is the basic environment configuration setup needed for the app to start
type EnvSpec struct {
	OtelGRPCEndpoint string `envconfig:"otel_grpc_endpoint"`
	OtelHTTPEndpoint string `envconfig:"otel_http_endpoint"`
	TracingEnabled   bool   `envconfig:"tracing_enabled" default:"true"`

	LogLevel string `envconfig:"log_level" default:"error"`

	Port    int    `envconfig:"port" default:"8080"`
	BaseURL string `envconfig:"base_url" default:""`

	ErrorUiUrl   string `envconfig:"error_ui_url" default:""`
	SupportEmail string `envconfig:"support_email" default:""`

	DirectoryApiUrl string `envconfig:"directory_api_url" required:"true"`
}

type Flags struct {
	ShowVersion bool
}

func NewFlags() *Flags {
	f := new(Flags)

	flag.BoolVar(&f.ShowVersion, "version", false, "Show the app version and exit")
	flag.Parse()

	return f
}
