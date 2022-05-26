REPO ?= 364554757/reloader
TAG ?= $(shell git rev-parse --abbrev-ref HEAD | sed -e 's/\//-/g')-$(shell git rev-parse --short HEAD)
RELEASE_TAG = $(shell git describe --abbrev=0 --tags)

install:
	kubectl apply -f deploy/deploy.yaml

uninstall:
	kubectl delete -f deploy/deploy.yaml

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o ./bin/reloader ./cmd/reloader.go
	docker build . -t $(REPO):$(TAG) -f deploy/Dockerfile
	docker push $(REPO):$(TAG)

release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o ./bin/reloader ./cmd/reloader.go
	docker build . -t $(REPO):$(RELEASE_TAG) -f deploy/Dockerfile
	docker push $(REPO):$(RELEASE_TAG)