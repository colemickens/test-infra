#!/usr/bin/env bash

set -euo pipefail
set -x

mkdir -p "cluster/secrets/"
secret="cluster/secrets/azure-ci-secret.yaml"

subscription_id="${SUBSCRIPTION_ID}"
tenant_id="${TENANT_ID}"
client_id="${SERVICE_PRINCIPAL_CLIENT_ID}"
client_secret="${SERVICE_PRINCIPAL_CLIENT_SECRET}"

oidc_client_id="${OIDC_CLIENT_ID}"
oidc_client_secret="${OIDC_CLIENT_SECRET}"

subscription_id="$(echo -n $subscription_id | base64)"
tenant_id="$(echo -n $tenant_id | base64)"
client_id="$(echo -n $client_id | base64)"
client_secret="$(echo -n $client_secret | base64)"
oidc_client_id="$(echo -n $oidc_client_id | base64)"
oidc_client_secret="$(echo -n $oidc_client_secret | base64)"

cat << EOF > "${secret}"
apiVersion: v1
kind: Secret
metadata:
  name: azure-ci
type: Opaque
data:
  SUBSCRIPTION_ID: $subscription_id
  TENANT_ID: $tenant_id
  SERVICE_PRINCIPAL_CLIENT_ID: $client_id
  SERVICE_PRINCIPAL_CLIENT_SECRET: $client_secret
  OIDC_CLIENT_ID: $oidc_client_id
  OIDC_CLIENT_SECRET: $oidc_client_secret
EOF

kubectl apply -f "${secret}"
