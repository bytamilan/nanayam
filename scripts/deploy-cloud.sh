#!/usr/bin/env bash
# =============================================================================
# Nanayam - one-command cloud deployment
# =============================================================================
# Builds the gateway and console images, pushes them to a registry, and applies
# the Kubernetes manifests in k8s/ to whatever cluster kubectl currently points
# at. Works against GKE, EKS, AKS, k3d/kind, or any conformant cluster.
#
#   ./scripts/deploy-cloud.sh --registry ghcr.io/bytamilan --domain nanayam.example.com
#
# Run with --dry-run first to see the rendered manifests without touching the
# cluster, and --destroy to remove everything it created.
# =============================================================================
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# --- Defaults ----------------------------------------------------------------
NAMESPACE="nanayam"
REGISTRY=""
IMAGE_TAG=""
DOMAIN=""
INGRESS_CLASS="nginx"
PROFILE="basic"
CHANNEL="mychannel"
CHAINCODE="basic"
MSP_ID="Org1MSP"
PEER_HOST="peer0.org1.nanayam.com"
PEER_PORT="7051"
CRYPTO_DIR=""
GATEWAY_REPLICAS="2"
CONSOLE_REPLICAS="2"
SIGNUP_ENABLED="false"
WAIT_TIMEOUT="5m"

SKIP_BUILD="false"
SKIP_PUSH="false"
DRY_RUN="false"
DESTROY="false"

# --- Output ------------------------------------------------------------------
if [[ -t 1 ]]; then
  BLUE=$'\033[34m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; RED=$'\033[31m'; BOLD=$'\033[1m'; RESET=$'\033[0m'
else
  BLUE=""; GREEN=""; YELLOW=""; RED=""; BOLD=""; RESET=""
fi

step() { printf '%s→%s %s\n' "$BLUE" "$RESET" "$*"; }
ok()   { printf '%s✓%s %s\n' "$GREEN" "$RESET" "$*"; }
warn() { printf '%s!%s %s\n' "$YELLOW" "$RESET" "$*" >&2; }
die()  { printf '%s✗%s %s\n' "$RED" "$RESET" "$*" >&2; exit 1; }

usage() {
  cat <<'USAGE'
Nanayam cloud deployment

Usage: ./scripts/deploy-cloud.sh [options]

Images and registry:
  --registry <ref>        Registry prefix for images, e.g. ghcr.io/bytamilan
                          Required unless --skip-build and --skip-push are set.
  --tag <tag>             Image tag (default: short git SHA, or "latest")
  --skip-build            Reuse images already in the registry
  --skip-push             Build locally without pushing (k3d/kind/minikube)

Cluster and routing:
  --namespace <ns>        Kubernetes namespace (default: nanayam)
  --domain <host>         Hostname for the ingress; omit to stay cluster-internal
  --ingress-class <name>  Ingress class (default: nginx)
  --gateway-replicas <n>  Gateway replica count (default: 2)
  --console-replicas <n>  Console replica count (default: 2)
  --wait-timeout <dur>    Rollout wait timeout (default: 5m)

Fabric wiring:
  --profile <basic|complaint>
                          Preset channel, chaincode, MSP, and peer (default: basic)
  --channel <name>        Override the channel name
  --chaincode <name>      Override the chaincode name
  --msp-id <id>           Override the MSP id
  --peer <host:port>      Override the peer endpoint
  --crypto-dir <path>     Fabric crypto material to upload as a secret
                          (default: ./crypto-config/peerOrganizations/<peer domain>)
  --enable-signup         Allow self-registration on the console (default: off)

Modes:
  --dry-run               Render manifests to stdout; change nothing
  --destroy               Delete the namespace and everything in it
  -h, --help              Show this help

Examples:
  # Deploy to a cloud cluster behind a hostname
  ./scripts/deploy-cloud.sh --registry ghcr.io/bytamilan --domain nanayam.example.com

  # Deploy the complaint network
  ./scripts/deploy-cloud.sh --registry ghcr.io/bytamilan --profile complaint

  # Local k3d cluster, images loaded directly, no registry
  ./scripts/deploy-cloud.sh --registry nanayam --skip-push

  # See what would be applied
  ./scripts/deploy-cloud.sh --registry ghcr.io/bytamilan --dry-run

  # Tear it all down
  ./scripts/deploy-cloud.sh --destroy
USAGE
}

# --- Argument parsing --------------------------------------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    --registry)          REGISTRY="${2:?--registry needs a value}"; shift 2 ;;
    --tag)               IMAGE_TAG="${2:?--tag needs a value}"; shift 2 ;;
    --namespace)         NAMESPACE="${2:?--namespace needs a value}"; shift 2 ;;
    --domain)            DOMAIN="${2:?--domain needs a value}"; shift 2 ;;
    --ingress-class)     INGRESS_CLASS="${2:?--ingress-class needs a value}"; shift 2 ;;
    --profile)           PROFILE="${2:?--profile needs a value}"; shift 2 ;;
    --channel)           CHANNEL="${2:?--channel needs a value}"; shift 2 ;;
    --chaincode)         CHAINCODE="${2:?--chaincode needs a value}"; shift 2 ;;
    --msp-id)            MSP_ID="${2:?--msp-id needs a value}"; shift 2 ;;
    --peer)              PEER_ENDPOINT_OVERRIDE="${2:?--peer needs a value}"; shift 2 ;;
    --crypto-dir)        CRYPTO_DIR="${2:?--crypto-dir needs a value}"; shift 2 ;;
    --gateway-replicas)  GATEWAY_REPLICAS="${2:?--gateway-replicas needs a value}"; shift 2 ;;
    --console-replicas)  CONSOLE_REPLICAS="${2:?--console-replicas needs a value}"; shift 2 ;;
    --wait-timeout)      WAIT_TIMEOUT="${2:?--wait-timeout needs a value}"; shift 2 ;;
    --enable-signup)     SIGNUP_ENABLED="true"; shift ;;
    --skip-build)        SKIP_BUILD="true"; shift ;;
    --skip-push)         SKIP_PUSH="true"; shift ;;
    --dry-run)           DRY_RUN="true"; shift ;;
    --destroy)           DESTROY="true"; shift ;;
    -h|--help)           usage; exit 0 ;;
    *)                   usage >&2; die "unknown option: $1" ;;
  esac
done

# --- Profile presets ---------------------------------------------------------
# Applied before any explicit override so --channel and friends still win.
case "$PROFILE" in
  basic) ;;
  complaint)
    [[ "$CHANNEL"   == "mychannel" ]] && CHANNEL="complaint-channel"
    [[ "$CHAINCODE" == "basic"     ]] && CHAINCODE="complaint"
    [[ "$MSP_ID"    == "Org1MSP"   ]] && MSP_ID="ACBMSP"
    [[ "$PEER_HOST" == "peer0.org1.nanayam.com" ]] && PEER_HOST="peer0.acb.nanayam.com"
    ;;
  *) die "unknown profile: $PROFILE (expected basic or complaint)" ;;
esac

if [[ -n "${PEER_ENDPOINT_OVERRIDE:-}" ]]; then
  PEER_HOST="${PEER_ENDPOINT_OVERRIDE%%:*}"
  if [[ "$PEER_ENDPOINT_OVERRIDE" == *:* ]]; then
    PEER_PORT="${PEER_ENDPOINT_OVERRIDE##*:}"
  fi
fi
PEER_ENDPOINT="${PEER_HOST}:${PEER_PORT}"

# The peer's org domain is everything after the leading "peerN." label.
PEER_DOMAIN="${PEER_HOST#*.}"
if [[ -z "$CRYPTO_DIR" ]]; then
  CRYPTO_DIR="${REPO_ROOT}/crypto-config/peerOrganizations/${PEER_DOMAIN}"
fi

# --- Preflight ---------------------------------------------------------------
require_tool() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required but not installed. $2"
}

preflight() {
  # --dry-run only renders text, so it must work with no tooling and no cluster.
  if [[ "$DRY_RUN" == "true" ]]; then
    return
  fi

  step "Checking prerequisites"
  require_tool kubectl "See https://kubernetes.io/docs/tasks/tools/"

  if true; then
    if ! kubectl cluster-info >/dev/null 2>&1; then
      die "kubectl cannot reach a cluster. Point your kubeconfig at the target cluster first:
  GKE: gcloud container clusters get-credentials <cluster> --region <region>
  EKS: aws eks update-kubeconfig --name <cluster> --region <region>
  AKS: az aks get-credentials --resource-group <rg> --name <cluster>
  k3d: k3d cluster create nanayam"
    fi
    ok "Cluster reachable: $(kubectl config current-context)"
  fi

  if [[ "$SKIP_BUILD" != "true" ]]; then
    require_tool docker "See https://docs.docker.com/get-docker/"
    ok "Docker available"
  fi
}

# --- Image build and push ----------------------------------------------------
resolve_tag() {
  if [[ -n "$IMAGE_TAG" ]]; then
    return
  fi
  IMAGE_TAG="$(git -C "$REPO_ROOT" rev-parse --short HEAD 2>/dev/null || echo latest)"
}

build_images() {
  if [[ "$SKIP_BUILD" == "true" ]]; then
    step "Skipping image build (--skip-build)"
    return
  fi

  step "Building gateway image: ${GATEWAY_IMAGE}"
  docker build -t "$GATEWAY_IMAGE" "${REPO_ROOT}/services/gateway"
  ok "Built ${GATEWAY_IMAGE}"

  step "Building console image: ${CONSOLE_IMAGE}"
  docker build -t "$CONSOLE_IMAGE" "${REPO_ROOT}/apps/org-console"
  ok "Built ${CONSOLE_IMAGE}"
}

push_images() {
  if [[ "$SKIP_PUSH" == "true" ]]; then
    step "Skipping image push (--skip-push)"
    warn "The cluster must already be able to resolve ${GATEWAY_IMAGE}."
    warn "For k3d: k3d image import ${GATEWAY_IMAGE} ${CONSOLE_IMAGE}"
    warn "For kind: kind load docker-image ${GATEWAY_IMAGE} ${CONSOLE_IMAGE}"
    return
  fi

  step "Pushing images to ${REGISTRY}"
  docker push "$GATEWAY_IMAGE"
  docker push "$CONSOLE_IMAGE"
  ok "Images pushed"
}

# --- Secrets -----------------------------------------------------------------
# The gateway needs the peer's MSP and TLS material. It is uploaded as a secret
# rather than baked into the image so the same image serves every organisation.
apply_crypto_secret() {
  if [[ ! -d "$CRYPTO_DIR" ]]; then
    die "Fabric crypto material not found at ${CRYPTO_DIR}
Generate it first with 'nanayam crypto generate' (or ./scripts/setup-fabric.sh),
or point at an existing directory with --crypto-dir."
  fi

  step "Uploading Fabric crypto material from ${CRYPTO_DIR}"
  kubectl create secret generic nanayam-fabric-crypto \
    --namespace "$NAMESPACE" \
    --from-file="$CRYPTO_DIR" \
    --dry-run=client -o yaml | kubectl apply -f -
  ok "Secret nanayam-fabric-crypto applied"
}

# The JWT secret is generated once and then left alone: regenerating it on every
# deploy would invalidate every signed-in session.
apply_auth_secret() {
  if kubectl get secret nanayam-auth --namespace "$NAMESPACE" >/dev/null 2>&1; then
    ok "Secret nanayam-auth already exists; leaving it in place"
    return
  fi

  step "Generating JWT signing secret"
  local secret
  secret="$(head -c 32 /dev/urandom | base64 | tr -d '\n')"
  kubectl create secret generic nanayam-auth \
    --namespace "$NAMESPACE" \
    --from-literal=jwt-secret="$secret" \
    --dry-run=client -o yaml | kubectl apply -f -
  ok "Secret nanayam-auth created"
}

# --- Manifest rendering ------------------------------------------------------
# Substitution is done with sed over an explicit variable list rather than
# envsubst, which is not installed everywhere and would also expand any unrelated
# $VAR that happens to appear in the manifests.
render_manifests() {
  local out="$1"
  : > "$out"

  local files=("${REPO_ROOT}/k8s/00-namespace.yaml"
               "${REPO_ROOT}/k8s/10-gateway.yaml"
               "${REPO_ROOT}/k8s/20-console.yaml")
  if [[ -n "$DOMAIN" ]]; then
    files+=("${REPO_ROOT}/k8s/30-ingress.yaml")
  fi

  local file
  for file in "${files[@]}"; do
    [[ -f "$file" ]] || die "manifest not found: $file"
    printf -- '---\n' >> "$out"
    sed \
      -e "s|\${NANAYAM_NAMESPACE}|${NAMESPACE}|g" \
      -e "s|\${NANAYAM_GATEWAY_IMAGE}|${GATEWAY_IMAGE}|g" \
      -e "s|\${NANAYAM_CONSOLE_IMAGE}|${CONSOLE_IMAGE}|g" \
      -e "s|\${NANAYAM_GATEWAY_REPLICAS}|${GATEWAY_REPLICAS}|g" \
      -e "s|\${NANAYAM_CONSOLE_REPLICAS}|${CONSOLE_REPLICAS}|g" \
      -e "s|\${NANAYAM_CHANNEL}|${CHANNEL}|g" \
      -e "s|\${NANAYAM_CHAINCODE}|${CHAINCODE}|g" \
      -e "s|\${NANAYAM_MSP_ID}|${MSP_ID}|g" \
      -e "s|\${NANAYAM_PEER_ENDPOINT}|${PEER_ENDPOINT}|g" \
      -e "s|\${NANAYAM_PEER_HOST}|${PEER_HOST}|g" \
      -e "s|\${NANAYAM_SIGNUP_ENABLED}|${SIGNUP_ENABLED}|g" \
      -e "s|\${NANAYAM_PUBLIC_GATEWAY_URL}|${PUBLIC_GATEWAY_URL}|g" \
      -e "s|\${NANAYAM_DOMAIN}|${DOMAIN}|g" \
      -e "s|\${NANAYAM_INGRESS_CLASS}|${INGRESS_CLASS}|g" \
      "$file" >> "$out"
  done
}

# --- Deploy and destroy ------------------------------------------------------
wait_for_rollout() {
  step "Waiting for rollout (timeout ${WAIT_TIMEOUT})"

  local failed=0
  local deployment
  for deployment in nanayam-gateway nanayam-console; do
    if kubectl rollout status "deployment/${deployment}" \
        --namespace "$NAMESPACE" --timeout "$WAIT_TIMEOUT"; then
      ok "${deployment} is ready"
    else
      warn "${deployment} did not become ready within ${WAIT_TIMEOUT}"
      failed=1
    fi
  done

  if [[ "$failed" -eq 1 ]]; then
    warn "Inspect the failing pods with:"
    warn "  kubectl -n ${NAMESPACE} get pods"
    warn "  kubectl -n ${NAMESPACE} logs -l app.kubernetes.io/part-of=nanayam --tail=100"
    return 1
  fi
}

print_access_instructions() {
  printf '\n%sNanayam is deployed.%s\n\n' "$BOLD" "$RESET"

  if [[ -n "$DOMAIN" ]]; then
    printf '  Console:  https://%s\n' "$DOMAIN"
    printf '  Gateway:  https://%s/v1\n' "$DOMAIN"
    printf '  Health:   https://%s/health\n\n' "$DOMAIN"
    printf '  DNS for %s must point at your ingress controller:\n' "$DOMAIN"
    printf '    kubectl get ingress -n %s\n\n' "$NAMESPACE"
  else
    printf '  No --domain was given, so the deployment is cluster-internal.\n'
    printf '  Reach it with port-forwarding:\n\n'
    printf '    kubectl -n %s port-forward svc/nanayam-console 3000:3000\n' "$NAMESPACE"
    printf '    kubectl -n %s port-forward svc/nanayam-gateway 8080:8080\n\n' "$NAMESPACE"
    printf '  Then open http://localhost:3000\n\n'
  fi

  printf '  Default sign-in is admin / admin. Change it before exposing this deployment.\n'
  printf '  Tear everything down with: ./scripts/deploy-cloud.sh --destroy --namespace %s\n\n' "$NAMESPACE"
}

destroy() {
  step "Deleting namespace ${NAMESPACE}"
  if ! kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    ok "Namespace ${NAMESPACE} does not exist; nothing to do"
    return
  fi

  kubectl delete namespace "$NAMESPACE" --wait=true
  ok "Namespace ${NAMESPACE} deleted"
}

main() {
  if [[ "$DESTROY" == "true" ]]; then
    require_tool kubectl "See https://kubernetes.io/docs/tasks/tools/"
    destroy
    return
  fi

  if [[ -z "$REGISTRY" ]]; then
    usage >&2
    die "--registry is required (for example: --registry ghcr.io/bytamilan)"
  fi

  resolve_tag
  GATEWAY_IMAGE="${REGISTRY}/nanayam-gateway:${IMAGE_TAG}"
  CONSOLE_IMAGE="${REGISTRY}/nanayam-console:${IMAGE_TAG}"

  if [[ -n "$DOMAIN" ]]; then
    PUBLIC_GATEWAY_URL="https://${DOMAIN}"
  else
    PUBLIC_GATEWAY_URL="http://localhost:8080"
  fi

  preflight

  local manifest
  manifest="$(mktemp -t nanayam-manifests.XXXXXX)"
  trap 'rm -f "$manifest"' EXIT
  render_manifests "$manifest"

  if [[ "$DRY_RUN" == "true" ]]; then
    step "Rendered manifests (--dry-run; nothing applied)"
    cat "$manifest"
    return
  fi

  build_images
  push_images

  # The namespace has to exist before the secrets that live in it, so it is
  # applied ahead of the rest of the rendered manifests.
  step "Applying namespace"
  sed -e "s|\${NANAYAM_NAMESPACE}|${NAMESPACE}|g" "${REPO_ROOT}/k8s/00-namespace.yaml" | kubectl apply -f -
  ok "Namespace ${NAMESPACE} ready"

  apply_crypto_secret
  apply_auth_secret

  step "Applying workloads"
  kubectl apply -f "$manifest"
  ok "Manifests applied"

  wait_for_rollout
  print_access_instructions
}

main "$@"
