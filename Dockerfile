FROM alpine
RUN apk add --no-cache ca-certificates
ADD views /app/views
ADD x9 /usr/bin/x9
WORKDIR /app
ENTRYPOINT ["/usr/bin/x9"]
