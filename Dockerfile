FROM golang:stretch

RUN go get github.com/google/uuid@v1.2.0
RUN go get github.com/deepmap/oapi-codegen/pkg/runtime

WORKDIR /app