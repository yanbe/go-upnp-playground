package ssdpsrv

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/huin/goupnp/httpu"
)

const (
	ntsAlive       = `ssdp:alive`
	ntsByebye      = `ssdp:byebye`
	ntsUpdate      = `ssdp:update`
	ssdpUDP4Addr   = "239.255.255.250:1900"
	ssdpSearchPort = 1900
	methodSearch   = "M-SEARCH"
	methodNotify   = "NOTIFY"

	// SSDPAll is a value for searchTarget that searches for all devices and services.
	SSDPAll = "ssdp:all"
	// UPNPRootDevice is a value for searchTarget that searches for all root devices.
	UPNPRootDevice = "upnp:rootdevice"
	server         = "go-upnp-playground/0.0.1"
	maxWaitSeconds = 1
)

// SSDPRawSearch performs a fairly raw SSDP search request, and returns the
// unique response(s) that it receives. Each response has the requested
// searchTarget, a USN, and a valid location. maxWaitSeconds states how long to
// wait for responses in seconds, and must be a minimum of 1 (the
// implementation waits an additional 100ms for responses to arrive), 2 is a
// reasonable value for this. numSends is the number of requests to send - 3 is
// a reasonable value for this.
func SSDPNotify() {
	uuid, _ := uuid.NewUUID()
	req := http.Request{
		Method: methodNotify,
		// TODO: Support both IPv4 and IPv6.
		Host: ssdpUDP4Addr,
		URL:  &url.URL{Opaque: "*"},
		Header: http.Header{
			// Putting headers in here avoids them being title-cased.
			// (The UPnP discovery protocol uses case-sensitive headers)
			"HOST":          []string{ssdpUDP4Addr},
			"CACHE-CONTROL": []string{"max-age = 1800"},
			"LOCATION":      []string{"http://192.168.10.10:8888/"},
			"NT":            []string{UPNPRootDevice},
			"NTS":           []string{ntsAlive},
			"USN":           []string{fmt.Sprintf("uuid:%s::urn:schemas-upnp-org:device:MediaServer:1", uuid)},
		},
	}
	client, _ := httpu.NewHTTPUClient()
	client.Do(&req, time.Duration(maxWaitSeconds)*time.Second+100*time.Millisecond, 2)
}
