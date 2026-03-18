# 跳过许可证检查
# FROM ghcr.io/openfaas/license-check:0.4.2 AS license-check

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.24 AS build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG VERSION
ARG GIT_COMMIT

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

# 跳过许可证检查
# COPY --from=license-check /license-check /usr/bin/

WORKDIR /go/src/github.com/openfaas/faas-netes
COPY . .

# 跳过许可证检查和测试
# RUN license-check -path /go/src/github.com/openfaas/faas-netes/ --verbose=false "Alex Ellis" "OpenFaaS Author(s)"
# RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*")
# RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go test -v ./...

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
        --ldflags "-s -w \
        -X github.com/openfaas/faas-netes/version.GitCommit=${GIT_COMMIT}\
        -X github.com/openfaas/faas-netes/version.Version=${VERSION}" \
        -o faas-netes .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.22.0 AS ship
# LABEL org.label-schema.license="OpenFaaS CE EULA - non-commercial" \
#       org.label-schema.vcs-url="https://github.com/openfaas/faas-netes" \
#       org.label-schema.vcs-type="Git" \
#       org.label-schema.name="openfaas/faas-netes" \
#       org.label-schema.vendor="openfaas" \
#       org.label-schema.docker.schema-version="1.0"

# 加一个这个，快一点
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && apk update

RUN apk --no-cache add \
    ca-certificates

RUN addgroup -S app \
    && adduser -S -g app app

WORKDIR /home/app

EXPOSE 8080

ENV http_proxy=""
ENV https_proxy=""

COPY --from=build /go/src/github.com/openfaas/faas-netes/faas-netes    .
RUN chown -R app:app ./

USER app

CMD ["./faas-netes"]
