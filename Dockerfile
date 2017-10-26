# build stage
FROM golang:1.8-alpine AS build-env
RUN apk add --no-cache git tar wget make
RUN wget -qO- https://github.com/Masterminds/glide/releases/download/v0.12.1/glide-v0.12.1-linux-amd64.tar.gz | tar xvz --strip-components=1 -C /go/bin/ linux-amd64/glide
ADD glide.yaml /go/src/github.com/vivareal/x9/glide.yaml
ADD glide.lock /go/src/github.com/vivareal/x9/glide.lock
WORKDIR /go/src/github.com/vivareal/x9
RUN glide install
ADD . /go/src/github.com/vivareal/x9
RUN make build
ENTRYPOINT ["/go/src/github.com/vivareal/x9/x9"]

# final stage
FROM alpine
RUN apk add --no-cache ca-certificates
ADD views /app/views
WORKDIR /app
COPY --from=build-env /go/src/github.com/vivareal/x9/x9 /usr/bin/x9
ENTRYPOINT ["/usr/bin/x9"]
