#!/bin/sh
set -u
image=${1:-}
requested_port=${2:-0}
host_address=127.0.0.1
container_port=8080
container_id=''
host_port=''
base_url=''
endpoint_attempts=3

require_command() {
	if command -v "$1" >/dev/null 2>&1; then
		return 0
	fi

	printf 'required command not found: %s\n' "$1" >&2
	return 127
}

cleanup() {
	if [ -n "$container_id" ]; then
		docker rm --force "$container_id" >/dev/null 2>&1 || :
	fi
}

report_failure() {
	printf '%s\n' 'container smoke test failed' >&2
	if [ -n "$container_id" ]; then
		docker inspect --format 'container status: {{.State.Status}}' "$container_id" >&2 || :
		docker logs "$container_id" >&2 || :
	fi
}

start_container() {
	if [ "$requested_port" = "0" ]; then
		if container_id=$(docker run --detach \
			--publish "${host_address}::${container_port}" \
			--env AGENT_PROVIDER=replay \
			"$image"); then
			:
		else
			return $?
		fi

		if host_port=$(docker port "$container_id" "${container_port}/tcp" |
			awk -F: 'NR == 1 {print $NF}'); then
			:
		else
			return $?
		fi
	else
		if container_id=$(docker run --detach \
			--publish "${host_address}:${requested_port}:${container_port}" \
			--env AGENT_PROVIDER=replay \
			"$image"); then
			:
		else
			return $?
		fi
		host_port=$requested_port
	fi

	if [ -n "$host_port" ]; then
		:
	else
		printf '%s\n' 'could not determine published container port' >&2
		return 1
	fi

	base_url="http://${host_address}:${host_port}"
	return 0
}

wait_for_readiness() {
	attempt=1
	while [ "$attempt" -le 30 ]; do
		if curl --fail --silent --max-time 2 \
			"${base_url}/readyz" >/dev/null; then
			return 0
		fi
		sleep 1
		attempt=$((attempt + 1))
	done

	report_failure
	return 1
}

validate_endpoints() {
	attempt=1
	while [ "$attempt" -le "$endpoint_attempts" ]; do
		if curl --fail --silent --max-time 2 \
			"${base_url}/healthz" >/dev/null &&
			curl --fail --silent --max-time 2 \
				"${base_url}/v1/state" \
				| jq -e '.availability == "available"' >/dev/null; then
			return 0
		fi
		sleep 1
		attempt=$((attempt + 1))
	done

	report_failure
	return 1
}

main() {
	if [ "$#" -lt 1 ] || [ "$#" -gt 2 ] || [ -z "$image" ]; then
		printf '%s\n' 'usage: smoke-container.sh IMAGE [HOST_PORT]' >&2
		return 2
	fi

	require_command docker || return $?
	require_command curl || return $?
	require_command jq || return $?

	trap cleanup 0 1 2 15

	start_container || return $?
	wait_for_readiness || return $?
	validate_endpoints || return $?

	printf '%s\n' 'container smoke test passed'
	return 0
}

main "$@"
