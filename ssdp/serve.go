package ssdp

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	ntsAlive       = `ssdp:alive`
	ntsByebye      = `ssdp:byebye`
	ntsUpdate      = `ssdp:update`
	ssdpUDP4Addr   = "239.255.255.250:1900"
	ssdpSearchPort = 1900
	methodNotify   = "NOTIFY"

	// SSDPAll is a value for searchTarget that searches for all devices and services.
	SSDPAll = "ssdp:all"
	// UPNPRootDevice is a value for searchTarget that searches for all root devices.
	UPNPRootDevice       = "upnp:rootdevice"
	UPNPContentDirectory = "urn:schemas-upnp-org:service:ContentDirectory:1"
	Vendor               = "Linux/i686 UPnP/1.0 go-upnp-playground/0.0.1"
	maxWaitSeconds       = 1
)

func NewServer(uuid uuid.UUID, deviceAddr *net.TCPAddr) Server {
	return Server{
		Addr:       ssdpUDP4Addr,
		Multicast:  true,
		uuid:       uuid,
		deviceAddr: deviceAddr,
	}
}

func (srv *Server) ServeMessage(rw http.ResponseWriter, req *http.Request) {
	mx, err := strconv.ParseInt(req.Header.Get("MX"), 10, 8)
	if err != nil {
		return
	}
	if mx > 120 {
		mx = 120
	}
	rand.Seed(time.Now().UnixNano())
	randSleepSeconds := rand.Intn(int(mx))
	log.Printf("[ssdp] Wait for random %d seconds for UPnP spec", randSleepSeconds)
	time.Sleep(time.Duration(randSleepSeconds) * time.Second)
	log.Printf("[ssdp] Wait for %d seconds finished", randSleepSeconds)

	res := http.Response{
		ProtoMajor: 1,
		ProtoMinor: 1,
		StatusCode: 200,
		Header: http.Header{
			// Putting headers in here avoids them being title-cased.
			// (The UPnP discovery protocol uses case-sensitive headers)
			"Cache-Control": {"max-age=1800"},
			"Location":      {fmt.Sprintf("http://%s/", srv.deviceAddr)},
			"Server":        {Vendor},
			"EXT":           {""},
			"USN":           {fmt.Sprintf("uuid:%s::%s", srv.uuid, UPNPRootDevice)},
			"ST":            {UPNPRootDevice},
			"Date":          {fmt.Sprintf("%v", time.Now().Format(time.RFC1123))},
		},
	}
	wb := bufio.NewWriter(rw)
	res.Write(wb)
	wb.Flush()
	log.Printf("[ssdp] Got %s %s message from %v: %v", req.Method, req.URL.Path, req.RemoteAddr, req.Header)

}

func (srv *Server) NotifyAlive() {
	req := http.Request{
		Method: methodNotify,
		// TODO: Support both IPv4 and IPv6.
		Host: ssdpUDP4Addr,
		URL:  &url.URL{Opaque: "*"},
		Header: http.Header{
			// Putting headers in here avoids them being title-cased.
			// (The UPnP discovery protocol uses case-sensitive headers)
			"Cache-Control": []string{"max-age=1800"},
			"Location":      []string{fmt.Sprintf("http://%s/", srv.deviceAddr)},
			"Server":        []string{"Linux/i686 UPnP/1.0 go-upnp-playground/0.0.1"},
			"NT":            []string{UPNPRootDevice},
			"NTS":           []string{ntsAlive},
			"USN":           {fmt.Sprintf("uuid:%s::%s", srv.uuid, UPNPRootDevice)},
		},
	}
	udpRoundTripper := UDPRoundTripper{}
	client := http.Client{Transport: &udpRoundTripper}
	client.Do(&req)
}

func (srv *Server) NotifyByebye() {
	req := http.Request{
		Method: methodNotify,
		// TODO: Support both IPv4 and IPv6.
		Host: ssdpUDP4Addr,
		URL:  &url.URL{Opaque: "*"},
		Header: http.Header{
			// Putting headers in here avoids them being title-cased.
			// (The UPnP discovery protocol uses case-sensitive headers)
			"NT":  []string{UPNPRootDevice},
			"NTS": []string{ntsByebye},
			"USN": {fmt.Sprintf("uuid:%s::%s", srv.uuid, UPNPRootDevice)},
		},
	}
	udpRoundTripper := UDPRoundTripper{}
	client := http.Client{Transport: &udpRoundTripper}
	client.Do(&req)
}
