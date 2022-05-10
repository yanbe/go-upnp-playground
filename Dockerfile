FROM golang:1.16-stretch

RUN go get github.com/google/uuid@v1.3.0
RUN go get github.com/deepmap/oapi-codegen/pkg/codegen@v1.10.1
RUN go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.10.1
RUN go get github.com/deepmap/oapi-codegen/pkg/runtime

WORKDIR /app
