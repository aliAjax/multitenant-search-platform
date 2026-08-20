#!/usr/bin/env bash
set -euo pipefail
base=${BASE_URL:-http://127.0.0.1:8086}
t=$(curl -sf -X POST "$base/v1/tenants" -H 'content-type: application/json' -d '{"name":"acme","quota":1000}')
tid=$(printf '%s' "$t" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
c=$(curl -sf -X POST "$base/v1/collections" -H 'content-type: application/json' -d "{\"tenant_id\":\"$tid\",\"name\":\"docs\",\"mappings\":{\"title\":{\"name\":\"title\",\"type\":\"text\"},\"kind\":{\"name\":\"kind\",\"type\":\"keyword\"}}}")
cid=$(printf '%s' "$c" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
curl -sf -X POST "$base/v1/documents" -H 'content-type: application/json' -d "{\"tenant_id\":\"$tid\",\"collection_id\":\"$cid\",\"data\":{\"title\":\"Go search platform\",\"kind\":\"guide\"}}"
curl -sf -X POST "$base/v1/search?collection_id=$cid" -H 'content-type: application/json' -d '{"match":{"title":"search"},"size":10}'
curl -sf -X POST "$base/v1/snapshots"
printf 'smoke ok\n'
