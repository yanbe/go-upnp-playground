package desc

import (
	"bytes"
	"log"
	"net"
	"net/http"
	"strconv"
	"text/template"

	"go-upnp-playground/soap"
)

var DeviceUUID string
var addr string
var xmlHeader = "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n"

// A Server defines parameters for running an HTTPU server.
type Server struct {
	listener net.Listener
}

func deviceDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[device] Got %s %s message from %v: %v", r.Method, r.URL.Path, r.RemoteAddr, r.Header)
	w.Header().Set("Content-Type", "text/xml")
	var buf = bytes.Buffer{}
	template.Must(template.ParseFiles("tmpl/device.xml")).Execute(&buf, map[string]interface{}{
		"uuid": DeviceUUID,
		"addr": addr,
	})
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}
func serviceContentDirectoryDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[device] Got %s %s message from %v: %v", r.Method, r.URL.Path, r.RemoteAddr, r.Header)
	w.Header().Set("CONTENT-TYPE", "text/xml")
	var buf = bytes.Buffer{}
	template.Must(template.ParseFiles("tmpl/ContentDirectory1.xml")).Execute(&buf, map[string]interface{}{})
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}
func serviceConnectionManagerDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[device] Got %s %s message from %v: %v", r.Method, r.URL.Path, r.RemoteAddr, r.Header)
	w.Header().Set("CONTENT-TYPE", "text/xml")
	var buf = bytes.Buffer{}
	template.Must(template.ParseFiles("tmpl/ConnectionManager1.xml")).Execute(&buf, map[string]interface{}{})
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}
func serviceContentDirectoryControlHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[device] Got %s %s message from %v: %v", r.Method, r.URL.Path, r.RemoteAddr, r.Header)
	w.Header().Set("Content-Type", `text/xml; charset="utf-8"`)
	w.Header().Set("Ext", "")
	w.Header().Set("Server", "Linux/i686 UPnP/1.0 go-upnp-playground/0.0.1")
	var buf = bytes.Buffer{}
	buf.WriteString(xmlHeader)
	buf.Write(soap.HandleAction(r))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}

func (srv *Server) Listen(hostIP string) string {
	srv.listener, _ = net.Listen("tcp", hostIP+":0")
	addr = srv.listener.Addr().(*net.TCPAddr).String()
	log.Println("Listening", addr)
	return addr
}

func (srv *Server) Serve() error {
	http.HandleFunc("/ContentDirectory/scpd.xml", serviceContentDirectoryDescriptionHandler)
	http.HandleFunc("/ContentDirectory/control.xml", serviceContentDirectoryControlHandler)
	http.HandleFunc("/ConnectionManager/scpd.xml", serviceConnectionManagerDescriptionHandler)
	http.HandleFunc("/ConnectionManager/control.xml", serviceContentDirectoryControlHandler)
	http.HandleFunc("/", deviceDescriptionHandler)
	return http.Serve(srv.listener, nil)
}

func NewServer() Server {
	return Server{}
}
