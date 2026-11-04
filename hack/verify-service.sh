#!/bin/sh
set -eu

: "${KUBE_CONTEXT:?set KUBE_CONTEXT to the target Kubernetes context}"
: "${KUBE_NAMESPACE:=anya-system}"
: "${EDGE_SERVICE:=anya-agent}"
: "${EDGE_LOCAL_PORT:=8080}"
: "${EDGE_REMOTE_PORT:=8080}"
host_address=127.0.0.1
base_url=''
port_forward_log=''
endpoint_attempts=3

require_command() {
	command -v "$1" >/dev/null 2>&1 || {
		printf 'required command not found: %s\n' "$1" >&2
		exit 127
	}
}

cleanup() {
	if [ -n "$port_forward_pid" ]; then
		kill "$port_forward_pid" 2>/dev/null || true
		wait "$port_forward_pid" 2>/dev/null || true
	fi
	if [ -n "$port_forward_log" ]; then
		rm -f "$port_forward_log"
	fi
}

start_port_forward() {
	port_forward_log=$(mktemp)
	kubectl --context "$KUBE_CONTEXT" \
		--namespace "$KUBE_NAMESPACE" \
		port-forward \
		"service/$EDGE_SERVICE" \
		"$EDGE_LOCAL_PORT:$EDGE_REMOTE_PORT" >"$port_forward_log" 2>&1 &
	port_forward_pid=$!
	base_url="http://${host_address}:${EDGE_LOCAL_PORT}"
}

wait_for_readiness() {
	readiness_attempt=1
	while [ "$readiness_attempt" -le "$EDGE_READY_TIMEOUT_SECONDS" ]; do
		if ! kill -0 "$port_forward_pid" 2>/dev/null; then
			printf '%s\n' 'port-forward exited unexpectedly' >&2
			cat "$port_forward_log" >&2
			return 1
		fi
		if curl --fail --silent --show-error --max-time 2 \
			"${base_url}/readyz" >/dev/null 2>&1; then
			return
		fi
		sleep 1
		readiness_attempt=$((readiness_attempt + 1))
	done

	printf '%s\n' 'agent did not become ready' >&2
	cat "$port_forward_log" >&2
	return 1
}

validate_agent_api() {
	for attempt in $(seq 1 "$endpoint_attempts"); do
		if validate_agent_api_once; then
			return
		fi
		sleep 1
	done

	printf '%s\n' 'service API validation failed' >&2
	return 1
}

validate_agent_api_once() {
	if ! curl --fail --silent --show-error --max-time 2 \
		"${base_url}/healthz" >/dev/null; then
		printf '%s\n' 'health endpoint failed' >&2
		return 1
	fi

	published_state_json=$(curl --fail --silent --show-error --max-time 2 \
		"${base_url}/v1/state") || return 1
	printf '%s\n' "$published_state_json" | jq -e '
		.availability == "available" and
		.health == "healthy" and
		(.snapshot.sequence > 0) and
		(.snapshot.identity.id != "") and
		(.snapshot.identity.model != "")
	' >/dev/null || return 1

	gpu_model=$(printf '%s\n' "$published_state_json" | jq -r '.snapshot.identity.model')
	sequence=$(printf '%s\n' "$published_state_json" | jq -r '.snapshot.sequence')
	health=$(printf '%s\n' "$published_state_json" | jq -r '.health')
	printf '      state: available, %s; GPU: %s; sequence: %s\n' \
		"$health" \
		"$gpu_model" \
		"$sequence"
}

main() {
	require_command kubectl
	require_command curl
	require_command jq
	trap cleanup EXIT INT TERM

	printf '%s\n' '[1/2] forward service and wait for readiness'
	start_port_forward
	wait_for_readiness
	printf '%s\n' '      passed'

	printf '%s\n' '[2/2] validate published service state'
	validate_agent_api
	printf '%s\n' '      passed'

	printf '%s\n' 'edge service smoke test passed'
}

main "$@"
