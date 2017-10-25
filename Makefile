HOME?=$$(HOME)
DOCKER_IMAGE_VERSION?=latest
DOCKER_REPO?=myrepo


build:
	CGO_ENABLED=0 go build -v -a --installsuffix cgo --ldflags="-s" -o x9

run:
	go run main.go

install:
	CGO_ENABLED=0 go install -v -a --installsuffix cgo --ldflags="-s"

docker_build:
	docker build -t ${DOCKER_REPO}/x9:build -f Dockerfile.build .

docker_image: docker_build
	docker run --rm --entrypoint /bin/sh -v ${PWD}:/out:rw ${DOCKER_REPO}/x9:build -c "cp /go/bin/x9 /out/x9"
	docker build -t ${DOCKER_REPO}/x9 .

docker_tag:
	docker tag ${DOCKER_REPO}/x9 ${DOCKER_REPO}/x9:${DOCKER_IMAGE_VERSION}

docker_run:
	docker run -d --rm ${DOCKER_REPO}/x9:${DOCKER_IMAGE_VERSION}

docker_push:
	docker push ${DOCKER_REPO}/x9:${DOCKER_IMAGE_VERSION}

