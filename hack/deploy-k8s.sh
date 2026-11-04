#!/bin/sh
set -eu

: "${KUBECTL_BIN:=kubectl}"
: "${KUBE_CONTEXT:?set KUBE_CONTEXT to the target Kubernetes context}"
: "${KUBE_IMAGE:?set KUBE_IMAGE to the immutable image to deploy}"
: "${KUBE_NAMESPACE:=anya-system}"

rendered_overlay_directory=''

require_command() {
	command -v "$1" >/dev/null 2>&1 || {
		printf 'required command not found: %s\n' "$1" >&2
		exit 127
	}
}

run_kubectl() {
	"$KUBECTL_BIN" --context "$KUBE_CONTEXT" "$@"
}

cleanup() {
	if [ -n "$rendered_overlay_directory" ]; then
		rm -rf "$rendered_overlay_directory"
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
	mkdir "$rendered_overlay_directory"
	cat >"$rendered_overlay_directory/kustomization.yaml" <<EOF
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
}

apply_manifests() {
	run_kubectl apply \
		--server-side \
		--namespace "$KUBE_NAMESPACE" \
		-k "$rendered_overlay_directory"

}

main() {
	require_command "$KUBECTL_BIN"
	trap cleanup EXIT INT TERM
	render_image_overlay
	apply_manifests
}

main "$@"
