package main

import (
	"fmt"
	"go-upnp-playground/desc"
	"go-upnp-playground/ssdp"

	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
)

func main() {
	deviceUUID := uuid.NewString()

	desc.DeviceUUID = deviceUUID
	descsrv := desc.NewServer()
	hostIP := os.Getenv("HOST_IP")
	if hostIP == "" {
		os.Exit(1)
	}
	addr := descsrv.Listen(hostIP)

	errDescSrv := make(chan error)
	go func() {
		errDescSrv <- descsrv.Serve()
	}()

	ssdpadv := ssdp.NewSSDPAdvertiser(deviceUUID, addr)
	ssdpres := ssdp.NewSSDPDiscoveryResponder(deviceUUID, addr)

	errSsdpRes := make(chan error)
	go func() {
		errSsdpRes <- ssdpres.ListenAndServe()
	}()
	ssdpadv.NotifyAlive()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ssdpadv.NotifyByebye()
		os.Exit(1)
	}()

	msgDescSrv := <-errDescSrv
	fmt.Println(msgDescSrv)
	msgSsdpRes := <-errSsdpRes
	fmt.Println(msgSsdpRes)
}
