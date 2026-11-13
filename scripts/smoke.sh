#!/usr/bin/env bash
set -euo pipefail

backend_url="${BACKEND_URL:-http://localhost:8080}"
frontend_url="${FRONTEND_URL:-http://localhost:3000}"

fail() {
	echo "FAIL: $*" >&2
	exit 1
}

check_status() {
	local url="$1"
	local expected="$2"
	local status
	status="$(curl -sS --retry 10 --retry-connrefused --retry-delay 1 -o /dev/null -w '%{http_code}' "$url")" ||
		fail "cannot reach $url"
	[[ "$status" == "$expected" ]] || fail "$url returned $status, want $expected"
	echo "ok $url $status"
}

check_status "$backend_url/healthz" 200
check_status "$backend_url/readyz" 200
check_status "$frontend_url/" 200
check_status "$frontend_url/readyz" 200
check_status "$frontend_url/some/deep/link" 200

ready_body="$(curl -sS "$backend_url/readyz")"
grep -q '"postgres":"ok"' <<<"$ready_body" || fail "postgres is not ready"
grep -q '"redis":"ok"' <<<"$ready_body" || fail "redis is not ready"
echo "ok readiness dependencies"

headers="$(curl -sS -D - -o /dev/null "$backend_url/healthz")"
grep -qi '^x-request-id:' <<<"$headers" || fail "missing X-Request-Id"
grep -qi '^x-trace-id:' <<<"$headers" || fail "missing X-Trace-Id"
echo "ok correlation headers"

problem_type="$(curl -sS -o /dev/null -w '%{content_type}' "$backend_url/missing")"
[[ "$problem_type" == "application/problem+json" ]] || fail "404 is not problem+json, got $problem_type"
echo "ok problem+json"

proxied_type="$(curl -sS -o /dev/null -w '%{content_type}' "$frontend_url/api/v1/missing")"
[[ "$proxied_type" == "application/problem+json" ]] || fail "frontend /api proxy did not reach the backend, got $proxied_type"
echo "ok frontend api proxy"

index_body="$(curl -sS "$frontend_url/")"
grep -qi 'id="root"' <<<"$index_body" || fail "frontend index is missing the root element"
echo "ok frontend index"

echo "smoke test passed"
