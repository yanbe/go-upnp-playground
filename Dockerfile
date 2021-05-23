FROM golang:stretch

ENV GOPATH=/go
WORKDIR /app
COPY . .
RUN go get
RUN go build
ENTRYPOINT [ "./go-upnp-playground" ]