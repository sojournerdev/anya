#!/bin/sh
set -u

: "${KUBECTL_BIN:=kubectl}"
KUBE_CONTEXT=${KUBE_CONTEXT:-}
KUBE_IMAGE=${KUBE_IMAGE:-}
: "${KUBE_NAMESPACE:=anya-system}"
rendered_overlay_directory=''

require_command() {
	if command -v "$1" >/dev/null 2>&1; then
		return 0
	fi

	printf 'required command not found: %s\n' "$1" >&2
	return 127
}

run_kubectl() {
	"$KUBECTL_BIN" --context "$KUBE_CONTEXT" "$@"
}

cleanup() {
	if [ -n "$rendered_overlay_directory" ]; then
		rm -rf "$rendered_overlay_directory" || :
	fi
}

render_image_overlay() {
	image_name=${KUBE_IMAGE%:*}
	image_tag=${KUBE_IMAGE##*:}
	image_revision=${image_tag#git-}
	if [ -z "$image_name" ] || [ "$image_name" = "$KUBE_IMAGE" ] ||
		[ -z "$image_revision" ] || [ "$image_tag" = "$image_revision" ]; then
		printf 'KUBE_IMAGE must include a name and git-<commit> tag: %s\n' "$KUBE_IMAGE" >&2
		return 1
	fi

	rendered_overlay_directory="deploy/k8s/overlays/.anya-render-$$"
	if mkdir "$rendered_overlay_directory"; then
		:
	else
		return $?
	fi

	if cat >"$rendered_overlay_directory/kustomization.yaml" <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../edge
commonAnnotations:
  org.opencontainers.image.revision: $image_revision
images:
  - name: anya
    newName: $image_name
    newTag: $image_tag
EOF
	then
		return 0
	else
		status=$?
		return "$status"
	fi
}

apply_manifests() {
	run_kubectl apply \
		--server-side \
		--namespace "$KUBE_NAMESPACE" \
		-k "$rendered_overlay_directory"

}

main() {
	[ -n "$KUBE_CONTEXT" ] && [ -n "$KUBE_IMAGE" ] || {
		printf '%s\n' 'KUBE_CONTEXT and KUBE_IMAGE must be set' >&2
		return 2
	}

	require_command "$KUBECTL_BIN" || return $?
	trap cleanup 0 1 2 15
	render_image_overlay || return $?
	apply_manifests || return $?
	return 0
}

main "$@"
