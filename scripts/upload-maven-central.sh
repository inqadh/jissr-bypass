#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <central-bundle.zip>" >&2
  exit 2
fi

bundle_path="$1"
if [[ ! -f "$bundle_path" ]]; then
  echo "Bundle not found: $bundle_path" >&2
  exit 2
fi
if [[ -z "${CENTRAL_TOKEN_USERNAME:-}" || -z "${CENTRAL_TOKEN_PASSWORD:-}" ]]; then
  echo "CENTRAL_TOKEN_USERNAME and CENTRAL_TOKEN_PASSWORD are required." >&2
  exit 2
fi

publishing_type="${CENTRAL_PUBLISHING_TYPE:-USER_MANAGED}"
if [[ "$publishing_type" != "USER_MANAGED" && "$publishing_type" != "AUTOMATIC" ]]; then
  echo "CENTRAL_PUBLISHING_TYPE must be USER_MANAGED or AUTOMATIC." >&2
  exit 2
fi

auth_token="$(
  printf '%s:%s' "$CENTRAL_TOKEN_USERNAME" "$CENTRAL_TOKEN_PASSWORD" |
    base64 |
    tr -d '\r\n'
)"
upload_url="https://central.sonatype.com/api/v1/publisher/upload?name=jissr-bypass-android&publishingType=$publishing_type"

deployment_id="$(
  curl --fail-with-body --silent --show-error \
    --request POST \
    --header "Authorization: Bearer $auth_token" \
    --form "bundle=@$bundle_path;type=application/octet-stream" \
    "$upload_url"
)"
deployment_id="$(printf '%s' "$deployment_id" | tr -d '\r\n')"
if [[ ! "$deployment_id" =~ ^[0-9a-fA-F-]{36}$ ]]; then
  echo "Central returned an unexpected deployment identifier." >&2
  exit 1
fi

echo "Central deployment: $deployment_id"
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  printf 'deployment_id=%s\n' "$deployment_id" >>"$GITHUB_OUTPUT"
fi

for _ in $(seq 1 60); do
  status_json="$(
    curl --fail-with-body --silent --show-error \
      --request POST \
      --header "Authorization: Bearer $auth_token" \
      "https://central.sonatype.com/api/v1/publisher/status?id=$deployment_id"
  )"
  state="$(printf '%s' "$status_json" | jq -r '.deploymentState')"
  echo "Central state: $state"
  case "$state" in
    VALIDATED|PUBLISHED)
      exit 0
      ;;
    FAILED)
      printf '%s\n' "$status_json" | jq .
      exit 1
      ;;
  esac
  sleep 10
done

echo "Timed out waiting for Central validation." >&2
exit 1
