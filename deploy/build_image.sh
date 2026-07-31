#!/usr/bin/env bash
# Build once per architecture, then copy tags to the secondary registry without
# rebuilding. The publish mode produces <arch>-<timestamp> and <arch>-latest
# tags plus the multi-architecture <timestamp> and latest manifests.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
VERSION_FILE="${REPO_ROOT}/backend/cmd/server/VERSION"
PROXY_URL="${HTTP_PROXY:-http://hkproxy.mindray.com:8080}"
NO_PROXY_VALUE="${NO_PROXY:-localhost,127.0.0.1,quay.globalapp.mindray.com,registry.apps.prd01.ocp.mindray.com}"
PRIMARY_IMAGE="${PRIMARY_IMAGE:-tamakiramimy/sub2api-kiro}"
SECONDARY_IMAGE="${SECONDARY_IMAGE:-quay.globalapp.mindray.com/prd/sub2api-kiro}"
VERSION="$(tr -d '\r\n' < "${VERSION_FILE}")"
PUBLISH=false

usage() {
    cat <<'EOF'
Usage: deploy/build_image.sh [options]

Options:
  --publish                 Build AMD64 and ARM64 once each, then publish.
  --version VERSION         Override VERSION (must be YYYYMMDD_HHMMSS).
  --primary-image IMAGE     Primary registry image; built exactly twice here.
  --secondary-image IMAGE   Registry receiving copied tags; never rebuilt.
  -h, --help                Show this help.

Environment:
  HTTP_PROXY / HTTPS_PROXY  Build proxy, defaults to hkproxy.mindray.com:8080.
  NO_PROXY                  Direct hosts; Quay is included by default.
  PRIMARY_IMAGE             Defaults to tamakiramimy/sub2api-kiro.
  SECONDARY_IMAGE           Defaults to quay.globalapp.mindray.com/prd/sub2api-kiro.

Tag layout:
  IMAGE:amd64-YYYYMMDD_HHMMSS  IMAGE:amd64-latest
  IMAGE:arm64-YYYYMMDD_HHMMSS  IMAGE:arm64-latest
  IMAGE:YYYYMMDD_HHMMSS        IMAGE:latest
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --publish)
            PUBLISH=true
            ;;
        --version)
            VERSION="$2"
            shift
            ;;
        --primary-image)
            PRIMARY_IMAGE="$2"
            shift
            ;;
        --secondary-image)
            SECONDARY_IMAGE="$2"
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage >&2
            exit 2
            ;;
    esac
    shift
done

if [[ ! "${VERSION}" =~ ^[0-9]{8}_[0-9]{6}$ ]]; then
    echo "VERSION must use YYYYMMDD_HHMMSS, got: ${VERSION}" >&2
    exit 2
fi

COMMIT="$(git -C "${REPO_ROOT}" rev-parse --short HEAD)"
BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
BUILD_ARGS=(
    --build-arg "VERSION=${VERSION}"
    --build-arg "COMMIT=${COMMIT}"
    --build-arg "DATE=${BUILD_DATE}"
    --build-arg "HTTP_PROXY=${PROXY_URL}"
    --build-arg "HTTPS_PROXY=${HTTPS_PROXY:-${PROXY_URL}}"
    --build-arg "ALL_PROXY=${ALL_PROXY:-${PROXY_URL}}"
    --build-arg "NO_PROXY=${NO_PROXY_VALUE}"
    --build-arg "http_proxy=${PROXY_URL}"
    --build-arg "https_proxy=${HTTPS_PROXY:-${PROXY_URL}}"
    --build-arg "all_proxy=${ALL_PROXY:-${PROXY_URL}}"
    --build-arg "no_proxy=${NO_PROXY_VALUE}"
    --build-arg "ALPINE_REPOSITORY=https://mirrors.aliyun.com/alpine"
    --build-arg "NPM_CONFIG_REGISTRY=https://registry.npmmirror.com"
    --build-arg "GOPROXY=https://mirrors.aliyun.com/goproxy/,direct"
    --build-arg "GOSUMDB=off"
)

if [[ "${PUBLISH}" != true ]]; then
    docker build \
        --tag "sub2api:${VERSION}" \
        --tag "sub2api:latest" \
        "${BUILD_ARGS[@]}" \
        --file "${REPO_ROOT}/Dockerfile" \
        "${REPO_ROOT}"
    exit 0
fi

build_architecture() {
    local architecture="$1"
    local image_tag="${PRIMARY_IMAGE}:${architecture}-${VERSION}"

    docker buildx build --progress=plain --platform "linux/${architecture}" --push \
        --provenance=false --sbom=false \
        --tag "${image_tag}" \
        --tag "${PRIMARY_IMAGE}:${architecture}-latest" \
        "${BUILD_ARGS[@]}" \
        --file "${REPO_ROOT}/Dockerfile" \
        "${REPO_ROOT}"
}

copy_architecture() {
    local architecture="$1"
    local source="${PRIMARY_IMAGE}:${architecture}-${VERSION}"
    local destination="${SECONDARY_IMAGE}:${architecture}-${VERSION}"

    docker pull --platform "linux/${architecture}" "${source}"
    docker tag "${source}" "${destination}"
    docker tag "${source}" "${SECONDARY_IMAGE}:${architecture}-latest"
    docker push "${destination}"
    docker push "${SECONDARY_IMAGE}:${architecture}-latest"
}

create_manifest() {
    local image="$1"

    docker buildx imagetools create \
        --tag "${image}:${VERSION}" \
        --tag "${image}:latest" \
        "${image}:amd64-${VERSION}" \
        "${image}:arm64-${VERSION}"
}

create_secondary_manifest() {
    local image="$1"

    # docker manifest create keeps local definitions. Remove stale local copies
    # so repeated releases can replace both the timestamp and latest manifests.
    DOCKER_CLI_EXPERIMENTAL=enabled docker manifest rm "${image}:${VERSION}" >/dev/null 2>&1 || true
    DOCKER_CLI_EXPERIMENTAL=enabled docker manifest rm "${image}:latest" >/dev/null 2>&1 || true
    DOCKER_CLI_EXPERIMENTAL=enabled docker manifest create \
        "${image}:${VERSION}" \
        "${image}:amd64-${VERSION}" \
        "${image}:arm64-${VERSION}"
    DOCKER_CLI_EXPERIMENTAL=enabled docker manifest push "${image}:${VERSION}"
    DOCKER_CLI_EXPERIMENTAL=enabled docker manifest create \
        "${image}:latest" \
        "${image}:amd64-${VERSION}" \
        "${image}:arm64-${VERSION}"
    DOCKER_CLI_EXPERIMENTAL=enabled docker manifest push "${image}:latest"
}

# These are the only two compilation steps in publish mode.
build_architecture amd64
build_architecture arm64
create_manifest "${PRIMARY_IMAGE}"

# Secondary registry receives copied architecture tags and a manifest only.
copy_architecture amd64
copy_architecture arm64
create_secondary_manifest "${SECONDARY_IMAGE}"

docker buildx imagetools inspect "${PRIMARY_IMAGE}:${VERSION}"
DOCKER_CLI_EXPERIMENTAL=enabled docker manifest inspect "${SECONDARY_IMAGE}:${VERSION}"
