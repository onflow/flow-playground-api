SHORT_COMMIT := $(shell git rev-parse --short HEAD)
IMAGE_URL := gcr.io/dl-flow/playground-api

.PHONY: generate
generate:
	GO111MODULE=on go generate ./...

.PHONY: test
test:
	GO111MODULE=on go test ./...

.PHONY: run
run:
	GO111MODULE=on go run server/server.go

.PHONY: docker-build
docker-build:
	DOCKER_BUILDKIT=1 docker build --ssh default -t gcr.io/dl-flow/playground-api:latest -t "gcr.io/dl-flow/playground-api:$(SHORT_COMMIT)" .

.PHONY: docker-push
docker-push:
	docker push gcr.io/dl-flow/playground-api:latest
	docker push "gcr.io/dl-flow/playground-api:$(SHORT_COMMIT)"


#----------------------------------------------------------------------
# CD COMMANDS
#----------------------------------------------------------------------

.PHONY: deploy-staging
deploy-staging: update-deployment-image apply-staging-files monitor-rollout

# Staging YAMLs must have 'staging' in their name.
.PHONY: apply-staging-files
apply-staging-files:
	kconfig=$$(uuidgen); \
	echo "$$KUBECONFIG_STAGING" > $$kconfig; \
	files=$$(find ${K8S_YAMLS_LOCATION} -type f \( -name "*.yml" -or -name "*.yaml" \) | grep staging); \
	echo "$$files" | xargs -I {} kubectl --kubeconfig=$$kconfig apply -f {}


.PHONY: deploy-production
deploy-production: update-deployment-image apply-production-files monitor-rollout

# Production YAMLs must have 'production' in their name.
.PHONY: apply-production-files
apply-production-files:
	kconfig=$$(uuidgen); \
	echo "$$KUBECONFIG_PRODUCTION_2" > $$kconfig; \
	files=$$(find ${K8S_YAMLS_LOCATION} -type f \( -name "*.yml" -or -name "*.yaml" \) | grep staging); \
	echo "$$files" | xargs -I {} kubectl --kubeconfig=$$kconfig apply -f {}

# Deployment YAMLs must have 'deployment' in their name.
.PHONY: update-deployment-image
update-deployment-image: CONTAINER=flow-playground-api
update-deployment-image:
	@files=$$(find ${K8S_YAMLS_LOCATION} -type f \( -name "*.yml" -or -name "*.yaml" \) | grep deployment); \
	for i in $$files; do \
		patched=`openssl rand -hex 8`; \
		kubectl patch -f $$i -p '{"spec":{"template":{"spec":{"containers":[{"name":"${CONTAINER}","image":"${IMAGE_URL}:${SHORT_COMMIT}"}]}}}}' --local -o yaml > $$patched; \
		mv -f $$patched $$i; \
	done

.PHONY: monitor-rollout
monitor-rollout:
	kubectl --kubeconfig=$$kconfig rollout status deployments.apps flow-playground-api-v1
