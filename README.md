# User Verification Service

[![CI](https://github.com/canonical/user-verification-service/actions/workflows/ci.yaml/badge.svg)](https://github.com/canonical/user-verification-service/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/canonical/user-verification-service/branch/main/graph/badge.svg?token=Aloh6MWghg)](https://codecov.io/gh/canonical/user-verification-service)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/canonical/user-verification-service/badge)](https://securityscorecards.dev/viewer/?platform=github.com&org=canonical&repo=user-verification-service)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-%23FE5196.svg)](https://conventionalcommits.org)

![GitHub Release](https://img.shields.io/github/v/release/canonical/user-verification-service)
[![Go Reference](https://pkg.go.dev/badge/github.com/canonical/user-verification-service.svg)](https://pkg.go.dev/github.com/canonical/user-verification-service)

This is the User Verification Service for the Canonical Identity Platform.

## Running the Service


### Environment variables

Code dealing with the environment variables resides
in [here](internal/config/specs.go) where each attribute has an annotation which
is the lowercase of the environment variable name.

At the moment the application is sourcing the following from the environment:

- `OTEL_GRPC_ENDPOINT` - needed if we want to use the OTel gRPC exporter for
  traces
- `OTEL_HTTP_ENDPOINT` - needed if we want to use the OTel HTTP exporter for
  traces (if gRPC is specified this gets unused)
- `TRACING_ENABLED` - switch for tracing, defaults to enabled (`true`)
- `LOG_LEVEL` - log level, defaults to `error`
- `PORT` - HTTP server port, defaults to `8080`
- `BASE_URL` - the base url that the application will be running on
- `SALESFORCE_CONSUMER_KEY` - the salesforce consumer key
- `SALESFORCE_CONSUMER_SECRET` - the salesforce consumer secret
- `SALESFORCE_DOMAIN` - the salesforce domain (e.g. `https://test.salesforce.com`)

### Container

To build the UI OCI image, you
need [rockcraft](https://canonical-rockcraft.readthedocs-hosted.com). To install
rockcraft run:

```shell
sudo snap install rockcraft --channel=latest/edge --classic
```

To build the image:

```shell
rockcraft pack
```

In order to run the produced image with docker:

```shell
# Import the image to Docker
sudo /snap/rockcraft/current/bin/skopeo --insecure-policy \
    copy oci-archive:./user-verification-service_1.22.1_amd64.rock \
    docker-daemon:localhost:32000/user-verification-service:registry
# Run the image
docker run -d \
  -it \
  --rm \
  -p 8080:8080 \
  --name user-verification-service \
  -e SALESFORCE_CONSUMER_KEY="consumer-key" \
  -e SALESFORCE_CONSUMER_SECRET="consumer-secret" \
  -e SALESFORCE_DOMAIN="https://test.salesforce.com" \
  localhost:32000/user-verification-service:registry start user-verification-service
```

### Try it out

To try the identity-platform user verification service, you can use the `make dev` command.

Please install `docker` and `docker-compose`.

You need to have a registered GitHub OAuth application to use for logging in.
To register a GitHub OAuth application:

1) Go to <https://github.com/settings/applications/new>. The application
   name and homepage URL do not matter, but the Authorization callback URL must
   be `http://localhost:4433/self-service/methods/oidc/callback/github`.
2) Generate a client secret
3) Create a file called `.env` on the root of the repository and paste your
   client credentials:

```shell
CLIENT_ID=<client_id>
CLIENT_SECRET=<client_secret>
```

From the root folder of the repository, run:
```shell
make dev
```

To test it, you can use a dummy app deployed on port 4446. Open your browser and go to:
http://127.0.0.1:4446/

## Security

Please see [SECURITY.md](https://github.com/canonical/user-verification-service/blob/main/SECURITY.md)
for guidelines on reporting security issues.
