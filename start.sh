#!/bin/bash

set -x
set -e

cleanup () {
  docker compose -f ./docker-compose.dev.yml  down
  exit
}

trap "cleanup" INT EXIT

# Start dependencies
docker compose -f ./docker-compose.dev.yml up  --wait --force-recreate --build --remove-orphans -d || true

# Start client app
HYDRA_CONTAINER_ID=$(docker ps -aqf "name=user-verification-service-hydra-1")
HYDRA_IMAGE=ghcr.io/canonical/hydra:2.2.0-canonical

CLIENT_RESULT=$(docker exec "$HYDRA_CONTAINER_ID" \
  hydra create client \
    --endpoint http://127.0.0.1:4445 \
    --name "OIDC App" \
    --grant-type authorization_code,refresh_token,urn:ietf:params:oauth:grant-type:device_code \
    --response-type code \
    --format json \
    --scope openid,profile,offline_access,email \
    --redirect-uri http://127.0.0.1:4446/callback)

CLIENT_ID=$(echo "$CLIENT_RESULT" | cut -d '"' -f4)
CLIENT_SECRET=$(echo "$CLIENT_RESULT" | cut -d '"' -f12)

docker stop oidc_client || true
docker rm oidc_client || true
docker run --network="host" -d --name=oidc_client --rm $HYDRA_IMAGE \
  exec hydra perform authorization-code \
  --endpoint http://localhost:4444 \
  --client-id $CLIENT_ID \
  --client-secret $CLIENT_SECRET \
  --scope openid,profile,email,offline_access \
  --no-open --no-shutdown --format json

DIRECTORY_API_URL="http://localhost" TRACING_ENABLED="false" go run . serve
