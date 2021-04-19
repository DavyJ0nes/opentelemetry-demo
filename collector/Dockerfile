FROM alpine:latest as build

WORKDIR /src
RUN apk --update add ca-certificates
RUN wget https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/download/v0.17.0/otelcontribcol_linux_amd64
RUN chmod +x ./otelcontribcol_linux_amd64

FROM alpine:latest

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /src/otelcontribcol_linux_amd64 /otelcontribcol
ENTRYPOINT ["/otelcontribcol"]
EXPOSE 55680 55679