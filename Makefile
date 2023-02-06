
DOCKER_USERNAME ?= ehlers320
APPLICATION_NAME ?= datadog-remote-adapter
GIT_HASH ?= $(shell git log --format="%h" -n 1)

test: 
	go test -v ./... 

build:
	docker build --tag ${DOCKER_USERNAME}/${APPLICATION_NAME}:${GIT_HASH} .
	docker tag ${DOCKER_USERNAME}/${APPLICATION_NAME}:${GIT_HASH} ${DOCKER_USERNAME}/${APPLICATION_NAME}:latest

push: build
	docker push ${DOCKER_USERNAME}/${APPLICATION_NAME}:${GIT_HASH}
	docker push ${DOCKER_USERNAME}/${APPLICATION_NAME}:latest

prometheus:
	./tests/prometheus/prometheus --config.file=tests/prometheus/prometheus.yml