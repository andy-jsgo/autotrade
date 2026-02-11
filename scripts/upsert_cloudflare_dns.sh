#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 3 ]]; then
  echo "Usage: $0 <cf_token> <zone_name> <record_name> [record_name...]" >&2
  exit 1
fi

CF_TOKEN="$1"
ZONE_NAME="$2"
shift 2
RECORDS=("$@")
TARGET_IP="72.62.247.57"

ZONE_ID=$(curl -sS -X GET "https://api.cloudflare.com/client/v4/zones?name=${ZONE_NAME}" \
  -H "Authorization: Bearer ${CF_TOKEN}" \
  -H "Content-Type: application/json" | jq -r '.result[0].id')

if [[ -z "${ZONE_ID}" || "${ZONE_ID}" == "null" ]]; then
  echo "Unable to resolve zone id for ${ZONE_NAME}" >&2
  exit 1
fi

for name in "${RECORDS[@]}"; do
  record_id=$(curl -sS -X GET "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records?type=A&name=${name}" \
    -H "Authorization: Bearer ${CF_TOKEN}" \
    -H "Content-Type: application/json" | jq -r '.result[0].id // empty')

  payload=$(jq -nc --arg type "A" --arg name "$name" --arg content "$TARGET_IP" '{type:$type,name:$name,content:$content,ttl:120,proxied:false}')

  if [[ -n "${record_id}" ]]; then
    curl -sS -X PUT "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records/${record_id}" \
      -H "Authorization: Bearer ${CF_TOKEN}" \
      -H "Content-Type: application/json" \
      --data "$payload" >/dev/null
    echo "updated ${name} -> ${TARGET_IP}"
  else
    curl -sS -X POST "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records" \
      -H "Authorization: Bearer ${CF_TOKEN}" \
      -H "Content-Type: application/json" \
      --data "$payload" >/dev/null
    echo "created ${name} -> ${TARGET_IP}"
  fi
done
