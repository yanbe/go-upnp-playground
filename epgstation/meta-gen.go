// go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=epgstation --generate types,client -o client.go http://$EPGSTATION/api/docs
// NOTE: To to re-generate EPGStation client code, you have to pass EPGStation's IP:Port as $EPGSTATION environment variable
// $ EPGSTATION=192.168.10.10:8888 go generate
package epgstation
