DOCKER_TAGS ?= latest

all: build

build:
	docker build -t bloomapi/gce-docker -f ./Dockerfile .
	$(foreach tag,$(DOCKER_TAGS), docker tag bloomapi/gce-docker bloomapi/gce-docker:$(tag) || exit 1;)

push: build
	$(foreach tag,$(DOCKER_TAGS), docker push bloomapi/gce-docker:$(tag) || exit 1;)