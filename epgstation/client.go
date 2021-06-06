//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=epgstation --generate types,client -o api.gen.go http://192.168.10.10:8888/api/docs
package epgstation
