package main

import (
	"fmt"
	"go-upnp-playground/service"
	"go-upnp-playground/ssdp"
	"log"

	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
)

func main() {
	deviceUUID := uuid.NewString()
	log.Print("deviceUUID: ", deviceUUID)

	service.DeviceUUID = deviceUUID
	server := service.NewServer()
	hostIP := os.Getenv("HOST_IP")
	if hostIP == "" {
		log.Fatal("HOST_IP environemnt variable not passed")
	}
	addr := server.Listen(hostIP)

	errSrv := make(chan error)
	go func() {
		errSrv <- server.Serve()
	}()

	ssdpadv := ssdp.NewSSDPAdvertiser(deviceUUID, addr)
	ssdpres := ssdp.NewSSDPDiscoveryResponder(deviceUUID, addr)

	errSsdpRes := make(chan error)
	errSsdpAdvRes := make(chan error)

	go func() {
		errSsdpRes <- ssdpres.ListenAndServe()
	}()
	go func() {
		errSsdpAdvRes <- ssdpadv.Serve()
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ssdpadv.NotifyByebye()
		os.Exit(1)
	}()

	msgSrv := <-errSrv
	fmt.Println(msgSrv)
	msgSsdpRes := <-errSsdpRes
	fmt.Println(msgSsdpRes)
	msgSsdpAdvRes := <-errSsdpAdvRes
	fmt.Println(msgSsdpAdvRes)
}
