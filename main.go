package main

import (
	"fmt"
	"go-upnp-playground/device"
	"go-upnp-playground/ssdp"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
)

func main() {
	uuid, _ := uuid.NewUUID()

	descsrv := device.NewServer(uuid)
	hostIP := os.Getenv("HOST_IP")
	descsrv.Listen(hostIP)
	errDescSrv := make(chan error)
	go func() {
		errDescSrv <- descsrv.Serve()
	}()

	ssdpsrv := ssdp.NewServer(uuid, descsrv.Addr)
	errSsdpSrv := make(chan error)
	go func() {
		errSsdpSrv <- ssdpsrv.ListenAndServe()
	}()
	ssdpsrv.NotifyAlive()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ssdpsrv.NotifyByebye()
		os.Exit(1)
	}()

	msgDescSrv := <-errDescSrv
	fmt.Println(msgDescSrv)
	msgSsdpSrv := <-errSsdpSrv
	fmt.Println(msgSsdpSrv)
}
