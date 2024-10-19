VERSION := $(shell git rev-parse HEAD)

swag:
	@swag init -d ./cmd,./internal

run:
	@KO_DATA_PATH=cmd/kodata/ VERSION=$(VERSION) go run cmd/main.go

publish:
	@sed -i "s/^VERSION=.*/VERSION=${VERSION}/" cmd/kodata/.env.prod
	@KO_DOCKER_REPO=TODO ko build cmd/main.go