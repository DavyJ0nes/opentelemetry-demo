FROM golang:1.15 As Builder

WORKDIR /go/src/github.com/davyj0nes/opentelemetry-demo/go-service
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -tags netgo --installsuffix netgo -o service main.go handlers.go router.go

# -- Main Image
FROM alpine:3.12

LABEL MAINTAINER=davyj0nes
LABEL OWNER=davyj0nes

# add root CA TLS certs
RUN apk --no-cache add ca-certificates

RUN adduser -D -s /bin/sh app

# Copy binary from builder image
COPY --from=Builder /go/src/github.com/davyj0nes/opentelemetry-demo/go-service/service /bin/service
RUN chmod a+x /bin/service

# Ensure not runnning as root user
USER app

CMD service
