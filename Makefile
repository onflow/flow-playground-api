SHORT_COMMIT := $(shell git rev-parse --short HEAD)
CONTAINER := flow-playground-api
IMAGE_URL := gcr.io/dl-flow/playground-api
K8S_YAMLS_LOCATION := ./k8s
KUBECONFIG := $(shell uuidgen)

.PHONY: generate
generate:
	GO111MODULE=on go generate ./...

.PHONY: test
test:
	GO111MODULE=on go test ./...

.PHONY: test-datastore
test-datastore:
	DATASTORE_EMULATOR_HOST=localhost:8081 FLOW_STORAGEBACKEND=datastore GO111MODULE=on go test ./...

.PHONY: run
run:
	FLOW_DEBUG=true FLOW_SESSIONCOOKIESSECURE=false GO111MODULE=on go run server/server.go

.PHONY: run-datastore
run-datastore:
	DATASTORE_EMULATOR_HOST=localhost:8081 FLOW_STORAGEBACKEND=datastore FLOW_DATASTORE_GCPPROJECTID=flow-developer-playground FLOW_DEBUG=true FLOW_SESSIONCOOKIESSECURE=false GO111MODULE=on go run server/server.go

.PHONY: docker-build
docker-build:
	DOCKER_BUILDKIT=1 docker build --ssh default -t gcr.io/dl-flow/playground-api:latest -t "gcr.io/dl-flow/playground-api:$(SHORT_COMMIT)" .

.PHONY: docker-push
docker-push:
	docker push gcr.io/dl-flow/playground-api:latest
	docker push "gcr.io/dl-flow/playground-api:$(SHORT_COMMIT)"

.PHONY: start-datastore-emulator
start-datastore-emulator:
	gcloud beta emulators datastore start --no-store-on-disk

#----------------------------------------------------------------------
# CD COMMANDS
#----------------------------------------------------------------------

.PHONY: deploy-staging
deploy-staging: update-deployment-image apply-staging-files monitor-rollout

# Staging YAMLs must have 'staging' in their name.
.PHONY: apply-staging-files
apply-staging-files:
	echo "$$KUBECONFIG_STAGING_2" > ${KUBECONFIG}; \
	files=$$(find ${K8S_YAMLS_LOCATION} -type f \( -name "*.yml" -or -name "*.yaml" \) | grep staging); \
	echo "$$files" | xargs -I {} kubectl --kubeconfig=${KUBECONFIG} apply -f {}


.PHONY: deploy-production
deploy-production: update-deployment-image apply-production-files monitor-rollout

# Production YAMLs must have 'production' in their name.
.PHONY: apply-production-files
apply-production-files:
	kconfig=$$(uuidgen); \
	echo "$$KUBECONFIG_PRODUCTION_2" > ${KUBECONFIG}; \
	files=$$(find ${K8S_YAMLS_LOCATION} -type f \( -name "*.yml" -or -name "*.yaml" \) | grep staging); \
	echo "$$files" | xargs -I {} kubectl --kubeconfig=${KUBECONFIG} apply -f {}

# Deployment YAMLs must have 'deployment' in their name.
.PHONY: update-deployment-image
update-deployment-image:
	@files=$$(find ${K8S_YAMLS_LOCATION} -type f \( -name "*.yml" -or -name "*.yaml" \) | grep deployment); \
	for i in $$files; do \
		patched=`openssl rand -hex 8`; \
		kubectl patch -f $$i -p '{"spec":{"template":{"spec":{"containers":[{"name":"${CONTAINER}","image":"${IMAGE_URL}:${SHORT_COMMIT}"}]}}}}' --local -o yaml > $$patched; \
		mv -f $$patched $$i; \
	done

.PHONY: monitor-rollout
monitor-rollout:
	kubectl --kubeconfig=${KUBECONFIG} rollout status deployments.apps flow-playground-api-v1
