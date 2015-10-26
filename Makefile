# Short name: Short name, following [a-zA-Z_], used all over the place.
# Some uses for short name:
# - Docker image name
# - Kubernetes service, rc, pod, secret, volume names
SHORT_NAME := example-todo

# SemVer with build information is defined in the SemVer 2 spec, but Docker
# doesn't allow +, so we use -.
VERSION := 0.0.1-$(shell date "+%Y%m%d%H%M%S")

# Docker Root FS
BINDIR := ./rootfs

# Legacy support for DEV_REGISTRY, plus new support for DEIS_REGISTRY.
DEV_REGISTRY ?= $(eval docker-machine ip deis):5000
DEIS_REGISTY ?= ${DEV_REGISTRY}

# Kubernetes-specific information for RC, Service, and Image.
RC := manifests/${SHORT_NAME}-rc.yaml
SVC := manifests/${SHORT_NAME}-service.yaml
REDIS_CLUSTER_SVC := manifests/${SHORT_NAME}-redis-cluster-service.yaml
REDIS_STANDALONE_SVC := manifests/${SHORT_NAME}-redis-standalone-service.yaml

# Docker image name
IMAGE := deis/${SHORT_NAME}:${VERSION}

all: docker-build docker-push

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s'
	mv example-todo rootfs/example-todo

# For cases where we're building from local
# We also alter the RC file to set the image name.
docker-build: build
	docker build --rm -t ${IMAGE} rootfs
	perl -pi -e "s|image: .+|image: \"${IMAGE}\"|g" ${RC}

# Push to a registry that Kubernetes can access.
docker-push:
	docker push ${IMAGE}

# Deploy is a Kubernetes-oriented target
deploy: kube-up

kube-up:
	kubectl create -f ${REDIS_CLUSTER_SVC}
	kubectl create -f ${REDIS_STANDALONE_SVC}
	kubectl create -f ${SVC}
	kubectl create -f ${RC}

kube-down:
	-kubectl delete -f ${RC}
	-kubectl delete -f ${SVC}
	-kubectl delete -f ${REDIS_CLUSTER_SVC}
	-kubectl delete -f ${REDIS_STANDALONE_SVC}

.PHONY: all build docker-build docker-push kube-up kube-down deploy
