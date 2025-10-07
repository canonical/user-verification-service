package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"

	"github.com/canonical/user-verification-service/internal/config"
	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/canonical/user-verification-service/internal/monitoring/prometheus"
	"github.com/canonical/user-verification-service/internal/salesforce"
	"github.com/canonical/user-verification-service/internal/tracing"
	"github.com/canonical/user-verification-service/pkg/web"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve starts the web server",
	Long:  `Launch the web application, list of environment variables is available in the readme`,
	Run: func(cmd *cobra.Command, args []string) {
		main()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serve() error {
	specs := new(config.EnvSpec)
	if err := envconfig.Process("", specs); err != nil {
		panic(fmt.Errorf("issues with environment sourcing: %s", err))
	}

	logger := logging.NewLogger(specs.LogLevel)
	defer logger.Sync()

	logger.Debugf("env vars: %v", specs)

	monitor := prometheus.NewMonitor("user-verification-service", logger)
	tracer := tracing.NewTracer(tracing.NewConfig(specs.TracingEnabled, specs.OtelGRPCEndpoint, specs.OtelHTTPEndpoint, logger))

	var sf salesforce.SalesforceAPI
	if specs.SalesforceEnabled {
		sf = salesforce.NewClient(
			specs.SalesforceDomain,
			specs.SalesforceConsumerKey,
			specs.SalesforceConsumerSecret,
			tracer,
			monitor,
			logger,
		)
	} else {
		logger.Warn("Using salesforce noop client, do not use this in production")
		sf = salesforce.NewNoopClient(tracer, monitor, logger)
	}

	router := web.NewRouter(specs.ErrorUiUrl, specs.SupportEmail, specs.ApiToken, specs.UiBaseURL, sf, tracer, monitor, logger)
	logger.Infof("Starting server on port %v", specs.Port)

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%v", specs.Port),
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	var serverError error
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Security().SystemStartup()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError = fmt.Errorf("server error: %w", err)
			c <- os.Interrupt
		}
	}()

	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	logger.Security().SystemShutdown()
	if err := srv.Shutdown(ctx); err != nil {
		serverError = fmt.Errorf("server shutdown error: %w", err)
	}

	return serverError
}

func main() {
	if err := serve(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}
