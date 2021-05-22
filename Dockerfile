FROM golang:stretch

ENV GOPATH=/go
WORKDIR /go-upnp-playground
COPY  go.* ./
COPY *.go ./
RUN go get
RUN go build
ENTRYPOINT [ "./main" ]