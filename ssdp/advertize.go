package ssdp

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

const (
	methodNotify = "NOTIFY"
	ssdpUDP4Addr = "239.255.255.250:1900"
	ntsAlive     = `ssdp:alive`
	ntsByebye    = `ssdp:byebye`
	ntsUpdate    = `ssdp:update`
	serverName   = "Linux/i686 UPnP/1.0 go-upnp-playground/0.0.1"
)

type SSDPAdvertiser struct {
	uuid string
	addr string
}

func (s *SSDPAdvertiser) RoundTrip(req *http.Request) (*http.Response, error) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	req.Write(&buf)
	destAddr, err := net.ResolveUDPAddr("udp", req.Host)
	if err != nil {
		return nil, err
	}
	conn.WriteTo(buf.Bytes(), destAddr)
	return &http.Response{}, nil
}

func NewSSDPAdvertiser(uuid string, addr string) SSDPAdvertiser {
	return SSDPAdvertiser{
		uuid: uuid,
		addr: addr,
	}
}

func (s *SSDPAdvertiser) NotifyAlive() {
	req := http.Request{
		Method: methodNotify,
		// TODO: Support both IPv4 and IPv6.
		Host: ssdpUDP4Addr,
		URL:  &url.URL{Opaque: "*"},
		Header: http.Header{
			// Putting headers in here avoids them being title-cased.
			// (The UPnP discovery protocol uses case-sensitive headers)
			"Cache-Control": []string{"max-age=1800"},
			"Location":      []string{fmt.Sprintf("http://%s/", s.addr)},
			"Server":        []string{serverName},
			"NT":            []string{upnpRootDevice},
			"NTS":           []string{ntsAlive},
			"USN":           {fmt.Sprintf("uuid:%s::%s", s.uuid, upnpRootDevice)},
		},
	}
	udpRoundTripper := UDPRoundTripper{}
	client := http.Client{Transport: &udpRoundTripper}
	client.Do(&req)
}

func (s *SSDPAdvertiser) NotifyByebye() {
	req := http.Request{
		Method: methodNotify,
		// TODO: Support both IPv4 and IPv6.
		Host: ssdpUDP4Addr,
		URL:  &url.URL{Opaque: "*"},
		Header: http.Header{
			// Putting headers in here avoids them being title-cased.
			// (The UPnP discovery protocol uses case-sensitive headers)
			"NT":  []string{upnpRootDevice},
			"NTS": []string{ntsByebye},
			"USN": {fmt.Sprintf("uuid:%s::%s", s.uuid, upnpRootDevice)},
		},
	}
	client := http.Client{Transport: s}
	client.Do(&req)
}
