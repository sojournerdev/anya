#!/bin/sh
set -eu

: "${EDGE_IMAGE:?set EDGE_IMAGE to an immutable image reference (for example, registry.example/anya:git-<commit>)}"
: "${EDGE_TARGET_PLATFORM:?set EDGE_TARGET_PLATFORM to the target platform (for example, linux/arm64)}"
: "${IMAGE_SOURCE:=}"

git_commit=''

require_command() {
	command -v "$1" >/dev/null 2>&1 || {
		printf 'required command not found: %s\n' "$1" >&2
		exit 127
	}
}

validate_image_reference() {
	case "$EDGE_IMAGE" in
		''|*' '*|*'\t'*|*'\n'*|*@*)
			printf 'EDGE_IMAGE must be a tagged image reference without a digest: %s\n' "$EDGE_IMAGE" >&2
			return 1
		;;
	esac

	image_name=${EDGE_IMAGE%:*}
	image_tag=${EDGE_IMAGE##*:}
	if [ "$image_name" = "$EDGE_IMAGE" ] || [ -z "$image_name" ] || [ -z "$image_tag" ]; then
		printf 'EDGE_IMAGE must include an image name and tag: %s\n' "$EDGE_IMAGE" >&2
		return 1
	fi
	case "$image_tag" in
		git-*) ;;
		*)
			printf 'EDGE_IMAGE tag must identify a Git revision (git-<commit>): %s\n' "$EDGE_IMAGE" >&2
			return 1
		;;
	esac
}

validate_git_state() {
	git_commit=$(git rev-parse HEAD) || {
		printf '%s\n' 'unable to determine the current Git revision' >&2
		return 1
	}
	if [ -n "$(git status --porcelain)" ]; then
		printf '%s\n' 'working tree must be clean before building a Git-tagged image' >&2
		return 1
	fi
	if [ "$image_tag" != "git-$git_commit" ]; then
		printf 'EDGE_IMAGE tag %s does not match current Git revision %s\n' \
			"$image_tag" "git-$git_commit" >&2
		return 1
	fi
}

build_image() {
	docker buildx build \
		--platform "$EDGE_TARGET_PLATFORM" \
		--file Containerfile \
		--build-arg "IMAGE_REVISION=$git_commit" \
		--build-arg "IMAGE_SOURCE=$IMAGE_SOURCE" \
		--tag "$EDGE_IMAGE" \
		--load \
		.
}

main() {
	require_command docker
	require_command git
	validate_image_reference
	validate_git_state
	printf '[1/1] build %s for %s\n' "$EDGE_IMAGE" "$EDGE_TARGET_PLATFORM"
	build_image
	printf '%s\n' 'image build passed'
}

main "$@"
