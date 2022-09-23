FROM golang:1.18-alpine AS builder

ARG version=0.0.0
ARG build_time=local
ARG gitsha=dirty

RUN apk add --no-cache git

ARG GOPRIVATE
ARG HUMANITEC_GOMOD_USER
ARG HUMANITEC_GOMOD_TOKEN

RUN git config --global url."https://${HUMANITEC_GOMOD_USER}:${HUMANITEC_GOMOD_TOKEN}@github.com".insteadOf "https://github.com"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# https://stackoverflow.com/questions/36279253/go-compiled-binary-wont-run-in-an-alpine-docker-container-on-ubuntu-host
ENV CGO_ENABLED=0
RUN GOOS=linux GOARCH=amd64 go build -ldflags "\
   -X humanitec.io/go-service-template/internal/version.Version=${version} \
   -X humanitec.io/go-service-template/internal/version.BuildTime=${build_time} \
   -X humanitec.io/go-service-template/internal/version.GitSHA=${gitsha} \
   " -o /opt/server/server ./cmd/server


FROM alpine:3 as final

WORKDIR /opt/server

COPY --from=builder /opt/server .

COPY ./config config
COPY ./openapi openapi

# TODO: Copy other static files if needed, e.g. migrations:
# COPY ./migrations migrations

ENTRYPOINT ["/opt/server/server"]