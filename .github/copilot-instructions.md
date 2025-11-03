# Copilot Instructions for user-verification-service
## Architecture
- Service verifies Canonical OAuth logins by checking Salesforce employment data; core entry is go/cobra CLI in cmd/serve.go (invoked via go run . serve).
- Config is centralized in internal/config/specs.go using envconfig tags (lowercase envvar names) and defaults such as port 8080 and TRACING_ENABLED=true.
- web.NewRouter builds the chi mux, wiring monitoring middleware, tracing.OpenTelemetry, and mounting modules from pkg/user_verification, pkg/ui, pkg/metrics, and pkg/status.
- pkg/user_verification/service.go defers to internal/salesforce/client.go which issues SOQL queries via k-capehart/go-salesforce; SALESFORCE_ENABLED toggles between real and noop clients.
- logging.NewLogger returns both service logs and security logger; serve() defers Sync() and emits SystemStartup/SystemShutdown via logger.Security().
## Request Flow
- POST /api/v0/verify expects a WebhookPayload { "email": string } and optionally applies AuthMiddleware if API_TOKEN is set; Authorization header must exactly match the token value (no Bearer prefix).
- Handlers reply with ORY Kratos style WebhookErrorResponse (see pkg/user_verification/handlers.go) so keep the structure if adding new error cases.
- Non-employee lookups must record security events via logger.Security().AuthzFailureNotEmployee with logging.WithRequest(r).
- Monitoring.NewMiddleware ResponseTime() labels routes by pattern (path parameters normalized to id); new routes on the mux inherit metrics automatically.
- Tracing.NewTracer reads OTEL_GRPC_ENDPOINT/OTEL_HTTP_ENDPOINT; disable tracing via TRACING_ENABLED=false (see start.sh for local defaults).
## Developer Workflow
- make dev spins up the full Hydra/Kratos stack via docker-compose.dev.yml, provisions an OAuth client, and runs go run . serve with useful local env defaults.
- make test runs go test ./... twice to collect cobertura JSON/coverage while filtering out generated mocks; prefer this target so coverage.out/test.json stay consistent.
- make mocks installs mockgen@v0.3.0 then executes go generate ./...; call it whenever interfaces change to refresh mocks under pkg/**/mock_*.go.
- make build produces the binary via go build -o app ./; CGO is disabled by default (CGO_ENABLED=0) to simplify deployment.
- rockcraft pack builds the OCI rock image described in rockcraft.yaml for release packaging.
## Conventions
- API modules expose NewAPI(...).RegisterEndpoints(*chi.Mux); follow this pattern and keep route registration centralised in pkg/web/router.go.
- Environment-aware URLs (BASE_URL, UI_BASE_URL, ERROR_UI_URL) should be normalized via parseBaseURL and ui.registrationURL to ensure trailing slashes and support email text.
- Internal interfaces live beside their packages (for example internal/salesforce/interfaces.go); tests rely on them for gomock generation, so define behaviour there when adding dependencies.
- Prometheus metrics are predefined in internal/monitoring/prometheus/prometheus.go; extend histogram/gauge vectors there and reuse Set*Metric helpers.
- Status endpoints at /api/v0/status and /api/v0/version use pkg/status/build.go to include version.Version and build metadata; update internal/version/const.go when releasing.
