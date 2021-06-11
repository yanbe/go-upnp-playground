package service

import (
	"encoding/xml"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"go-upnp-playground/bufferpool"
	"go-upnp-playground/epgstation"
	"go-upnp-playground/service/contentdirectory"
	"go-upnp-playground/soap"

	"github.com/google/uuid"
)

func serveXMLDocHandler(tmplFile string, vars map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		if len(vars) == 0 {
			r, err := os.Open(tmplFile)
			if err != nil {
				log.Fatal("error on open file: ", err)
			}
			fi, err := r.Stat()
			if err != nil {
				log.Fatal("error on stat file: ", err)
			}
			w.Header().Set("Content-Length", strconv.Itoa(int(fi.Size())))
			io.Copy(w, r)
		} else {
			buf := bufferpool.NewBytesBuffer()
			defer bufferpool.PutBytesBuffer(buf)
			template.Must(template.ParseFiles(tmplFile)).Execute(buf, vars)
			w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
			w.Write(buf.Bytes())
		}
	}
}

func serviceContentDirectoryControlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", `text/xml; charset="utf-8"`)
	w.Header().Set("Ext", "")
	w.Header().Set("Server", "Linux/i686 UPnP/1.0 go-upnp-playground/0.0.1")
	buf := bufferpool.NewBytesBuffer()
	defer bufferpool.PutBytesBuffer(buf)
	buf.WriteString(xml.Header)
	buf.Write(soap.HandleAction(r))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}

func recordedVideoStreamHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.Query().Get("id")
	w.Header().Set("Content-Type", `text/xml; charset="utf-8"`)
	w.Header().Set("Server", "Linux/i686 UPnP/1.0 go-upnp-playground/0.0.1")
	buf := bufferpool.NewBytesBuffer()
	defer bufferpool.PutBytesBuffer(buf)
	buf.WriteString(xml.Header)
	buf.Write(soap.HandleAction(r))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}

func liveVideoStreamHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.Query().Get("id")
	w.Header().Set("Content-Type", `text/xml; charset="utf-8"`)
	w.Header().Set("Server", "Linux/i686 UPnP/1.0 go-upnp-playground/0.0.1")
	buf := bufferpool.NewBytesBuffer()
	defer bufferpool.PutBytesBuffer(buf)
	buf.WriteString(xml.Header)
	buf.Write(soap.HandleAction(r))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}

// A Server defines parameters for running an HTTPU server.
type Server struct {
	deviceUUID uuid.UUID
	listener   *net.TCPListener
	addr       net.TCPAddr
}

func (s *Server) Addr() net.TCPAddr {
	return s.addr
}

func (s *Server) Setup(hostIP net.IP) {
	s.addr.IP = hostIP

	addr := net.TCPAddr{}
	addr.IP, addr.Port = hostIP, 8888
	epgstation.Setup(addr)

	contentdirectory.Setup()

	http.HandleFunc("/", serveXMLDocHandler("tmpl/device.xml", map[string]interface{}{
		"uuid": s.deviceUUID,
		"addr": &s.addr,
	}))
	http.HandleFunc("/ContentDirectory/scpd.xml", serveXMLDocHandler("tmpl/ContentDirectory1.xml", nil))
	http.HandleFunc("/ConnectionManager/scpd.xml", serveXMLDocHandler("tmpl/ConnectionManager1.xml", nil))

	http.HandleFunc("/ContentDirectory/control.xml", serviceContentDirectoryControlHandler)
	http.HandleFunc("/ConnectionManager/control.xml", serviceContentDirectoryControlHandler)

	http.HandleFunc("/Streams/recorded", recordedVideoStreamHandler)
	http.HandleFunc("/Streams/live", liveVideoStreamHandler)
}

func (s *Server) Listen() {
	laddr := net.TCPAddr{}
	laddr.IP = s.addr.IP
	laddr.Port = 0
	var err error
	s.listener, err = net.ListenTCP("tcp", &laddr) // start listen arbitorary port
	if err != nil {
		log.Fatal(err)
	}
	s.addr.Port = s.listener.Addr().(*net.TCPAddr).Port
}

func (s *Server) Serve() error {
	return http.Serve(s.listener, nil)
}

func NewServer(deviceUUID uuid.UUID) *Server {
	return &Server{
		deviceUUID: deviceUUID,
		listener:   nil,
	}
}
