#!/bin/sh
set -u

KUBE_CONTEXT=${KUBE_CONTEXT:-}
: "${KUBE_NAMESPACE:=anya-system}"
: "${EDGE_SERVICE:=anya-agent}"
: "${EDGE_LOCAL_PORT:=8080}"
: "${EDGE_REMOTE_PORT:=8080}"
host_address=127.0.0.1
base_url=''
port_forward_pid=''
port_forward_log=''
endpoint_attempts=3

require_command() {
	if command -v "$1" >/dev/null 2>&1; then
		return 0
	fi

	printf 'required command not found: %s\n' "$1" >&2
	return 127
}

cleanup() {
	if [ -n "$port_forward_pid" ]; then
		kill "$port_forward_pid" 2>/dev/null || :
		wait "$port_forward_pid" 2>/dev/null || :
	fi
	if [ -n "$port_forward_log" ]; then
		rm -f "$port_forward_log" || :
	fi
}

start_port_forward() {
	if port_forward_log=$(mktemp); then
		:
	else
		return $?
	fi

	kubectl --context "$KUBE_CONTEXT" \
		--namespace "$KUBE_NAMESPACE" \
		port-forward \
		"service/$EDGE_SERVICE" \
		"$EDGE_LOCAL_PORT:$EDGE_REMOTE_PORT" >"$port_forward_log" 2>&1 &
	port_forward_pid=$!
	base_url="http://${host_address}:${EDGE_LOCAL_PORT}"
	return 0
}

wait_for_readiness() {
	readiness_attempt=1
	while [ "$readiness_attempt" -le "$EDGE_READY_TIMEOUT_SECONDS" ]; do
		if ! kill -0 "$port_forward_pid" 2>/dev/null; then
			printf '%s\n' 'port-forward exited unexpectedly' >&2
			cat "$port_forward_log" >&2 || :
			return 1
		fi
		if curl --fail --silent --show-error --max-time 2 \
			"${base_url}/readyz" >/dev/null 2>&1; then
			return 0
		fi
		sleep 1
		readiness_attempt=$((readiness_attempt + 1))
	done
	printf '%s\n' 'agent did not become ready' >&2
	cat "$port_forward_log" >&2 || :
	return 1
}

validate_agent_api() {
	attempt=1
	while [ "$attempt" -le "$endpoint_attempts" ]; do
		if validate_agent_api_once; then
			return 0
		fi
		sleep 1
		attempt=$((attempt + 1))
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

	if published_state_json=$(curl --fail --silent --show-error --max-time 2 \
		"${base_url}/v1/state"); then
		:
	else
		return $?
	fi
	if printf '%s\n' "$published_state_json" | jq -e '
		.availability == "available" and
		.health == "healthy" and
		(.snapshot.sequence > 0) and
		(.snapshot.identity.id != "") and
		(.snapshot.identity.model != "")
	' >/dev/null; then
		:
	else
		return $?
	fi

	if gpu_model=$(printf '%s\n' "$published_state_json" | jq -r '.snapshot.identity.model'); then
		:
	else
		return $?
	fi
	if sequence=$(printf '%s\n' "$published_state_json" | jq -r '.snapshot.sequence'); then
		:
	else
		return $?
	fi
	if health=$(printf '%s\n' "$published_state_json" | jq -r '.health'); then
		:
	else
		return $?
	fi
	printf '      state: available, %s; GPU: %s; sequence: %s\n' \
		"$health" \
		"$gpu_model" \
		"$sequence"
}

main() {
	[ -n "$KUBE_CONTEXT" ] || {
		printf '%s\n' 'KUBE_CONTEXT must be set' >&2
		return 2
	}

	require_command kubectl || return $?
	require_command curl || return $?
	require_command jq || return $?
	trap cleanup 0 1 2 15

	printf '%s\n' '[1/2] forward service and wait for readiness'
	start_port_forward || return $?
	wait_for_readiness || return $?
	printf '%s\n' '      passed'

	printf '%s\n' '[2/2] validate published service state'
	validate_agent_api || return $?
	printf '%s\n' '      passed'

	printf '%s\n' 'edge service smoke test passed'
	return 0
}

main "$@"
