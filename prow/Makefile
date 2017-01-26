# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

all: build fmt vet test


HOOK_VERSION   = 0.82
LINE_VERSION   = 0.73
SINKER_VERSION = 0.4
DECK_VERSION   = 0.16
SPLICE_VERSION = 0.13
MARQUE_VERSION = 0.1
TOT_VERSION    = 0.0
CRIER_VERSION  = 0.5


# These are the usual GKE variables.
PROJECT = k8s-prow
ZONE = us-central1-f
CLUSTER = prow

# These are the usual Azure variables.
LOCATION = westus
RESOURCE_GROUP = colemick-k8s-prow

# Cloud-specfic variables
#REGISTRY = $(REGISTRY)
#DOCKER   = gcloud docker
REGISTRY = docker.io/colemickens
DOCKER   = docker

LE_EMAIL   = cole.mickens@gmail.com
LE_DOMAINS = test0.prow.azure-containers.io
# TODO: sort this out, at least rename
# TODO: Maintainer in the image is a Googler

# These are GitHub credentials in files on your own machine.
# The hook secret is your HMAC token, the OAuth secret is the OAuth
# token of whatever account you want to comment and update statuses.
HOOK_SECRET_FILE = cluster/secrets/hook
OAUTH_SECRET_FILE = cluster/secrets/k8s-oauth-token

# The Jenkins secret is the API token, and the address file contains Jenkins'
# URL, such as http://pull-jenkins-master:8080, without a newline.
JENKINS_SECRET_FILE = cluster/secrets/jenkins
JENKINS_ADDRESS_FILE = cluster/secrets/jenkins-address

# Service account key for bootstrap jobs.
SERVICE_ACCOUNT_FILE = cluster/secrets/service-account.json

# Should probably move this to a script or something.
create-cluster-gce:
	gcloud -q container --project "$(PROJECT)" clusters create "$(CLUSTER)" --zone "$(ZONE)" --machine-type n1-standard-4 --num-nodes 4 --node-labels=role=prow --scopes "https://www.googleapis.com/auth/compute","https://www.googleapis.com/auth/devstorage.full_control","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management" --network "default" --enable-cloud-logging --enable-cloud-monitoring
	gcloud -q container node-pools create build-pool --project "$(PROJECT)" --cluster "$(CLUSTER)" --zone "$(ZONE)" --machine-type n1-standard-8 --num-nodes 4 --local-ssd-count=1 --node-labels=role=build

create-cluster-azure:
	az group get --name=$(RESOURCE_GROUP) \
		|| az group create --name=$(RESOURCE_GROUP) --location=$(LOCATION)
	az acs get --name=prow --resource-group=$(RESOURCE_GROUP) \
		|| az acs create --name=prow --resource-group=$(RESOURCE_GROUP) --orchestrator=kubernetes --location=$(LOCATION) --agent-count=10 --agent-vm-size=Standard_D5_v2

configure-cluster-azure:
	kubectl label node --all role=build role=prow --overwrite

deploy-prow:
	kubectl create secret generic hmac-token --from-file=hmac=$(HOOK_SECRET_FILE)
	kubectl create secret generic oauth-token --from-file=oauth=$(OAUTH_SECRET_FILE)
	kubectl create secret generic jenkins-token --from-file=jenkins=$(JENKINS_SECRET_FILE)
	kubectl create secret generic service-account --from-file=service-account.json=$(SERVICE_ACCOUNT_FILE)
	kubectl create configmap jenkins-address --from-file=jenkins-address=$(JENKINS_ADDRESS_FILE)
	kubectl create configmap config --from-file=config=config.yaml
	kubectl create configmap plugins --from-file=plugins=plugins.yaml
	@make update-prow

update-prow-ingress:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/ingress.yaml \
	| sed 's|__PROW_DOMAIN__|$(LE_DOMAINS)|' \
		| kubectl apply -f -

update-prow-nginx:
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/contrib/master/ingress/controllers/nginx/examples/default/rc-default.yaml
	kubectl apply -f https://gist.githubusercontent.com/colemickens/f5a1c751ca5ae54f0d6747895da062bc/raw/85c008b290e86e58508fee06b50afbd4648d313c/gistfile1.txt
	kubectl apply -f https://gist.githubusercontent.com/colemickens/8ef928fb3d642929cdc77dd67ce4dda9/raw/0aa652e2481467081479fb016d8d24b0295d3238/gistfile1.txt

update-prow:
	@make line-image --no-print-directory
	@make hook-image --no-print-directory
	@make sinker-image --no-print-directory
	@make deck-image --no-print-directory
	@make hook-deployment --no-print-directory
	@make sinker-deployment --no-print-directory
	@make deck-deployment --no-print-directory
	@make splice-image --no-print-directory
	@make splice-deployment --no-print-directory
	@make marque-image --no-print-directory
	@make marque-deployment --no-print-directory
	@make update-prow-ingress --no-print-directory

clean-prow:
	kubectl delete deployments marque deck hook sinker splice --ignore-not-found
	kubectl delete ingress ing --ignore-not-found
	kubectl delete configmap job-configs plugins --ignore-not-found
	kubectl delete secret hmac-token oauth-token --ignore-not-found
	kubectl delete job --all --ignore-not-found

update-jobs:
	kubectl create configmap config --from-file=config=config.yaml --dry-run -o yaml | kubectl replace configmap config -f -

update-plugins:
	kubectl create configmap plugins --from-file=plugins=plugins.yaml --dry-run -o yaml | kubectl replace configmap plugins -f -

build:
	go install ./cmd/...

test:
	go test -race -cover $$(go list ./... | grep -v "\/vendor\/")

.PHONY: create-cluster update-cluster update-jobs update-plugins clean build test get-cluster-credentials

hook-image:
	CGO_ENABLED=0 go build -o cmd/hook/hook k8s.io/test-infra/prow/cmd/hook
	docker build -t "$(REGISTRY)/hook:$(HOOK_VERSION)" cmd/hook
	$(DOCKER) -- push "$(REGISTRY)/hook:$(HOOK_VERSION)"

hook-deployment:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/hook_deployment.yaml | kubectl apply -f -

hook-service:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/hook_service.yaml | kubectl apply -f -

line-image:
	CGO_ENABLED=0 go build -o cmd/line/line k8s.io/test-infra/prow/cmd/line
	docker build -t "$(REGISTRY)/line:$(LINE_VERSION)" cmd/line
	$(DOCKER) -- push "$(REGISTRY)/line:$(LINE_VERSION)"

sinker-image:
	CGO_ENABLED=0 go build -o cmd/sinker/sinker k8s.io/test-infra/prow/cmd/sinker
	docker build -t "$(REGISTRY)/sinker:$(SINKER_VERSION)" cmd/sinker
	$(DOCKER) -- push "$(REGISTRY)/sinker:$(SINKER_VERSION)"

sinker-deployment:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/sinker_deployment.yaml | kubectl apply -f -

deck-image:
	CGO_ENABLED=0 go build -o cmd/deck/deck k8s.io/test-infra/prow/cmd/deck
	docker build -t "$(REGISTRY)/deck:$(DECK_VERSION)" cmd/deck
	$(DOCKER) -- push "$(REGISTRY)/deck:$(DECK_VERSION)"

deck-deployment:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/deck_deployment.yaml | kubectl apply -f -

deck-service:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/deck_service.yaml | kubectl apply -f -

splice-image:
	CGO_ENABLED=0 go build -o cmd/splice/splice k8s.io/test-infra/prow/cmd/splice
	docker build -t "$(REGISTRY)/splice:$(SPLICE_VERSION)" cmd/splice
	$(DOCKER) -- push "$(REGISTRY)/splice:$(SPLICE_VERSION)"

splice-deployment:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/splice_deployment.yaml | kubectl apply -f -

marque-image:
	CGO_ENABLED=0 go build -o cmd/marque/marque k8s.io/test-infra/prow/cmd/marque
	docker build -t "$(REGISTRY)/marque:$(MARQUE_VERSION)" cmd/marque
	$(DOCKER) -- push "$(REGISTRY)/marque:$(MARQUE_VERSION)"

marque-deployment:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/marque_deployment.yaml \
	| sed 's|__LETSENCRYPT_EMAIL__|$(LE_EMAIL)|g' \
	| sed 's|__LETSENCRYPT_DOMAINS__|$(LE_DOMAINS)|g' \
		| kubectl apply -f -

marque-service:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/marque_service.yaml | kubectl apply -f -

tot-image:
	CGO_ENABLED=0 go build -o cmd/tot/tot k8s.io/test-infra/prow/cmd/tot
	docker build -t "$(REGISTRY)/tot:$(TOT_VERSION)" cmd/tot
	$(DOCKER) -- push "$(REGISTRY)/tot:$(TOT_VERSION)"

tot-deployment:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/tot_deployment.yaml | kubectl apply -f -

tot-service:
	sed 's|__REGISTRY__|$(REGISTRY)|g' cluster/tot_service.yaml | kubectl apply -f -

crier-image:
	CGO_ENABLED=0 go build -o cmd/crier/crier k8s.io/test-infra/prow/cmd/crier
	docker build -t "gcr.io/$(PROJECT)/crier:$(CRIER_VERSION)" cmd/crier
	gcloud docker -- push "gcr.io/$(PROJECT)/crier:$(CRIER_VERSION)"

crier-deployment:
	kubectl apply -f cluster/crier_deployment.yaml

crier-service:
	kubectl apply -f cluster/crier_service.yaml

.PHONY: hook-image hook-deployment hook-service test-pr-image sinker-image sinker-deployment deck-image deck-deployment deck-service splice-image splice-deployment marque-image marque-deployment marque-service tot-image tot-service tot-deployment crier-image crier-service crier-deployment
