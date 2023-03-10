#DOCKER_IMAGE="posbus-client"
#DOCKER_TAG="develop"

all: build

build:
	go build -trimpath -o ./bin/pbc ./cmd/standalone
	#tinygo build -o ./bin/pbc ./cmd/standalone

worker:
	GOOS=js GOARCH=wasm go build -trimpath -o ./bin/worker.wasm ./cmd/worker

run: build
	./bin/pbc

pbupdate:
	GOPROXY=direct go get -u github.com/momentum-xyz/ubercontroller/pkg/posbus@develop && go mod vendor
#test:
#	go test -v -race ./...
#

test:
	go build -trimpath -o ./bin/test ./cmd/test && ./bin/test



#build-docs:
#	swag init -g api.go -d universe/node,./,universe/streamchat -o build/docs/

#docker-build: DOCKER_BUILDKIT=1
#docker-build:
#	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
#
## docker run ...
#docker: docker-build
#	docker run --rm ${DOCKER_IMAGE}:${DOCKER_TAG}
#
#.PHONY: build run test docker docker-build
