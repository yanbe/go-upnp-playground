package desc

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"text/template"

	"go-upnp-playground/soap"
)

var DeviceUUID string
var addr string

// A Server defines parameters for running an HTTPU server.
type Server struct {
	listener net.Listener
}

func deviceDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	wb := bufio.NewWriter(w)
	template.Must(template.ParseFiles("tmpl/device.xml")).Execute(wb, map[string]interface{}{
		"uuid": DeviceUUID,
		"addr": addr,
	})
	wb.Flush()
	log.Printf("[device] Got %s %s message from %v: %v", r.Method, r.URL.Path, r.RemoteAddr, r.Header)
}
func serviceDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	wb := bufio.NewWriter(w)
	template.Must(template.ParseFiles("tmpl/service.xml")).Execute(wb, map[string]interface{}{})
	wb.Flush()
	log.Printf("[device] Got %s %s message from %v: %v", r.Method, r.URL.Path, r.RemoteAddr, r.Header)
}
func serviceControlHandler(w http.ResponseWriter, r *http.Request) {
	res := soap.HandleAction(r)
	log.Print(string(res))
	wb := bufio.NewWriter(w)
	wb.Write(res)
	wb.Flush()
	log.Printf("[device] Got %s %s message from %v: %v", r.Method, r.URL.Path, r.RemoteAddr, r.Header)
}

func (srv *Server) Listen(hostIP string) string {
	srv.listener, _ = net.Listen("tcp", hostIP+":0")
	addr = srv.listener.Addr().(*net.TCPAddr).String()
	log.Println("Listening", addr)
	return addr
}

func (srv *Server) Serve() error {
	http.HandleFunc("/ContentDirectory/scpd.xml", serviceDescriptionHandler)
	http.HandleFunc("/ContentDirectory/control.xml", serviceControlHandler)
	http.HandleFunc("/", deviceDescriptionHandler)
	return http.Serve(srv.listener, nil)
}

func NewServer() Server {
	return Server{}
}
