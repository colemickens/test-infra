#!/usr/bin/env bash

set -euo pipefail
set -x

mkdir -p "cluster/secrets"
OAUTH_SECRET="${OAUTH_SECRET:-cluster/secrets/k8s-oauth-token}"
HMAC_SECRET="${HMAC_SECRET:-cluster/secrets/hook}"

if [[ ! -f "${OAUTH_SECRET}" ]]; then
	echo "generating personal access token (oauth token) on github -> ${OAUTH_SECRET}"
	set +x

	pat_name="@acs-bot: (generated: $(hostname)-$(date +"%Y%m%d-%H%M%S"))"
	if [[ -z "${GITHUB_USERNAME:-}" ]]; then echo "enter GITHUB_USERNAME:"; read GITHUB_USERNAME; fi
	if [[ -z "${GITHUB_PASSWORD:-}" ]]; then echo "enter GITHUB_PASSWORD:"; read GITHUB_PASSWORD; fi
	if [[ -z "${GITHUB_OTP:-}" ]]; then echo "enter GITHUB_OTP:"; read GITHUB_OTP; fi
	
	resp="$(curl \
		-f \
		-u "${GITHUB_USERNAME}:${GITHUB_PASSWORD}" \
		-H "X-GitHub-OTP: ${GITHUB_OTP}" \
		-H "Accept: Accept: application/vnd.github.v3+json" \
		--data "{\"scopes\": [ \"public_repo\" ], \"note\": \"${pat_name}\" }" \
		https://api.github.com/authorizations)"

	echo "${resp}" | jq -r '.token' > "${OAUTH_SECRET}"
	set -x
fi


if [[ ! -f "${HMAC_SECRET}" ]]; then
	echo "generating hmac secret -> ${HMAC_SECRET}"
	set +x
	uuidgen > "${HMAC_SECRET}"
	set -x
fi
