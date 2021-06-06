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
	upnpRootDevice        = "upnp:rootdevice"
	upnpMediaServer       = "urn:schemas-upnp-org:device:MediaServer:1"
	upnpContentDirectory  = "urn:schemas-upnp-org:service:ContentDirectory:1"
	upnpConnectionManager = "urn:schemas-upnp-org:service:ConnectionManager:1"
	vendor                = "Linux/i686 UPnP/1.0 go-upnp-playground/0.0.1"
	maxWaitSeconds        = 1
)

func NewSSDPDiscoveryResponder(uuid string, addr string) SSDPDiscoveryResponder {
	return SSDPDiscoveryResponder{
		Multicast: true,
		uuid:      uuid,
		addr:      addr,
	}
}

func (s *SSDPDiscoveryResponder) stAndUSN(target string) (string, string) {
	var ST string
	var USN string
	deviceTarget := fmt.Sprintf("uuid:%s", s.uuid)
	switch target {
	case deviceTarget:
		ST = deviceTarget
		USN = ST
	case upnpRootDevice:
		fallthrough
	case upnpContentDirectory:
		fallthrough
	case upnpConnectionManager:
		ST = target
		USN = fmt.Sprintf("%s::%s", deviceTarget, ST)
	}
	return ST, USN
}

func waitRandomMillis(mx int64) {
	rand.Seed(time.Now().UnixNano())
	randSleepMilliSeconds := rand.Intn(int(mx))
	log.Printf("mx: %d, random wait: %d", mx, randSleepMilliSeconds)
	time.Sleep(time.Duration(randSleepMilliSeconds) * time.Millisecond)
}

func (srv *SSDPDiscoveryResponder) ServeMessage(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "M-SEARCH" {
		return
	}
	ST, USN := srv.stAndUSN(req.Header.Get("ST"))
	if ST == "" {
		return
	}
	log.Printf("[ssdp] Got %s %s message from %v: %v", req.Method, req.URL.Path, req.RemoteAddr, req.Header)
	mxHeader := req.Header.Get("MX")
	mx, err := strconv.ParseInt(mxHeader, 10, 8)
	if err != nil {
		log.Fatalf("invalid MX header: cannot parse as int %s", mxHeader)
	}
	if mx > 120 {
		mx = 120
	}
	waitRandomMillis(mx * 1000)
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
			"USN":           {USN},
			"ST":            {ST},
			"Date":          {fmt.Sprintf("%v", time.Now().Format(time.RFC1123))},
		},
	}
	log.Printf("[device] Respond %v", res)
	wb := bufio.NewWriter(rw)
	res.Write(wb)
	wb.Flush()
}
