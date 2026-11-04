#!/bin/sh
set -eu
image=${1:?usage: smoke-container.sh IMAGE [HOST_PORT]}
requested_port=${2:-0}
host_address=127.0.0.1
container_port=8080
container_id=''
host_port=''
base_url=''
endpoint_attempts=3

require_command() {
	command -v "$1" >/dev/null 2>&1 || {
		printf 'required command not found: %s\n' "$1" >&2
		exit 127
	}
}

cleanup() {
	if [ -n "$container_id" ]; then
		docker rm --force "$container_id" >/dev/null 2>&1 || true
	fi
}

report_failure() {
	printf '%s\n' 'container smoke test failed' >&2
	if [ -n "$container_id" ]; then
		docker inspect --format 'container status: {{.State.Status}}' "$container_id" >&2 || true
		docker logs "$container_id" >&2 || true
	fi
}

start_container() {
	if [ "$requested_port" = "0" ]; then
		container_id=$(docker run --detach \
			--publish "${host_address}::${container_port}" \
			--env AGENT_PROVIDER=replay \
			"$image")
		host_port=$(docker port "$container_id" "${container_port}/tcp" | awk -F: 'NR == 1 {print $NF}')
	else
		container_id=$(docker run --detach \
			--publish "${host_address}:${requested_port}:${container_port}" \
			--env AGENT_PROVIDER=replay \
			"$image")
		host_port=$requested_port
	fi
	base_url="http://${host_address}:${host_port}"
}

wait_for_readiness() {
	for attempt in $(seq 1 30); do
		if curl --fail --silent --max-time 2 \
			"${base_url}/readyz" >/dev/null; then
			return
		fi
		sleep 1
	done

	report_failure
	exit 1
}

validate_endpoints() {
	for attempt in $(seq 1 "$endpoint_attempts"); do
		if curl --fail --silent --max-time 2 \
			"${base_url}/healthz" >/dev/null &&
			curl --fail --silent --max-time 2 \
				"${base_url}/v1/state" \
				| jq -e '.availability == "available"' >/dev/null; then
			return
		fi
		sleep 1
	done

	report_failure
	exit 1
}

main() {
	require_command docker
	require_command curl
	require_command jq
	trap cleanup EXIT INT TERM
	start_container
	wait_for_readiness
	validate_endpoints
	printf '%s\n' 'container smoke test passed'
}

main "$@"
