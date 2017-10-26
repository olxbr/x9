HOME?=$$(HOME)

build:
	CGO_ENABLED=0 go build -v -a --installsuffix cgo --ldflags="-s" -o x9

run:
	go run main.go

install:
	CGO_ENABLED=0 go install -v -a --installsuffix cgo --ldflags="-s"

docker_build:
	docker-compose build

docker_run:
	docker-compose up -d

docker_push:
	docker-compose push
