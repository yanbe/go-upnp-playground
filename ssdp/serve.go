package ssdp

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	// upnpRootDevice is a value for searchTarget that searches for all root devices.
	upnpRootDevice       = "upnp:rootdevice"
	upnpContentDirectory = "urn:schemas-upnp-org:service:ContentDirectory:1"
	vendor               = "Linux/i686 UPnP/1.0 go-upnp-playground/0.0.1"
	maxWaitSeconds       = 1
)

func NewSSDPDiscoveryResponder(uuid string, addr string) SSDPDiscoveryResponder {
	return SSDPDiscoveryResponder{
		Multicast: true,
		uuid:      uuid,
		addr:      addr,
	}
}

func (srv *SSDPDiscoveryResponder) ServeMessage(rw http.ResponseWriter, req *http.Request) {
	// Devices should wait a random interval less than 100 milliseconds before sending an initial set of advertisements in order to
	// reduce the likelihood of network storms
	mx, err := strconv.ParseInt(req.Header.Get("MX"), 10, 8)
	if err != nil {
		return
	}
	if mx > 120 {
		mx = 120
	}
	rand.Seed(time.Now().UnixNano())
	randSleepSeconds := rand.Intn(int(mx))
	time.Sleep(time.Duration(randSleepSeconds) * time.Second)

	res := http.Response{
		ProtoMajor: 1,
		ProtoMinor: 1,
		StatusCode: 200,
		Header: http.Header{
			// Putting headers in here avoids them being title-cased.
			// (The UPnP discovery protocol uses case-sensitive headers)
			"Cache-Control": {"max-age=1800"},
			"Location":      {fmt.Sprintf("http://%s/", srv.addr)},
			"Server":        {vendor},
			"EXT":           {""},
			"USN":           {fmt.Sprintf("uuid:%s::%s", srv.uuid, upnpRootDevice)},
			"ST":            {upnpRootDevice},
			"Date":          {fmt.Sprintf("%v", time.Now().Format(time.RFC1123))},
		},
	}
	wb := bufio.NewWriter(rw)
	res.Write(wb)
	wb.Flush()
	log.Printf("[ssdp] Got %s %s message from %v: %v", req.Method, req.URL.Path, req.RemoteAddr, req.Header)
}
