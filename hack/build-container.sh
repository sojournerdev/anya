#!/bin/sh
set -u

IMAGE=${IMAGE:-}
: "${IMAGE_SOURCE:=}"
: "${IMAGE_REVISION:=}"
: "${IMAGE_VERSION:=}"
: "${IMAGE_CREATED:=}"

require_command() {
	if command -v "$1" >/dev/null 2>&1; then
		return 0
	fi

	printf 'required command not found: %s\n' "$1" >&2
	return 127
}

build_image() {
	set -- docker buildx build
	if [ -n "${TARGET_PLATFORM:-}" ]; then
		set -- "$@" --platform "$TARGET_PLATFORM"
	fi
	set -- "$@" \
		--file Containerfile \
		--build-arg "IMAGE_REVISION=$IMAGE_REVISION" \
		--build-arg "IMAGE_SOURCE=$IMAGE_SOURCE" \
		--build-arg "IMAGE_VERSION=$IMAGE_VERSION" \
		--build-arg "IMAGE_CREATED=$IMAGE_CREATED" \
		--tag "$IMAGE" \
		--load \
		.
	"$@"
}

main() {
	if [ "$#" -ne 0 ]; then
		printf '%s\n' 'usage: build-container.sh' >&2
		return 2
	fi
	[ -n "$IMAGE" ] || {
		printf '%s\n' 'IMAGE must be set' >&2
		return 2
	}
	require_command docker || return $?
	build_image || return $?
	printf 'image build passed: %s\n' "$IMAGE"
}

main "$@"
