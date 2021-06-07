package service

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"go-upnp-playground/soap"
)

var DeviceUUID string
var addr string

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
			var buf = bytes.Buffer{}
			template.Must(template.ParseFiles(tmplFile)).Execute(&buf, vars)
			w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
			w.Write(buf.Bytes())
		}
	}
}

func serviceContentDirectoryControlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", `text/xml; charset="utf-8"`)
	w.Header().Set("Ext", "")
	w.Header().Set("Server", "Linux/i686 UPnP/1.0 go-upnp-playground/0.0.1")
	var buf = bytes.Buffer{}
	buf.WriteString(xml.Header)
	buf.Write(soap.HandleAction(r))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}

// A Server defines parameters for running an HTTPU server.
type Server struct {
	listener net.Listener
}

func (srv *Server) Listen(hostIP string) string {
	srv.listener, _ = net.Listen("tcp", hostIP+":0") // start listen arbitorary port
	addr = srv.listener.Addr().(*net.TCPAddr).String()
	log.Println("Listening", addr)
	return addr
}

func (srv *Server) Serve() error {
	http.HandleFunc("/", serveXMLDocHandler("tmpl/device.xml", map[string]interface{}{
		"uuid": DeviceUUID,
		"addr": addr,
	}))
	http.HandleFunc("/ContentDirectory/scpd.xml", serveXMLDocHandler("tmpl/ContentDirectory1.xml", nil))
	http.HandleFunc("/ContentDirectory/control.xml", serviceContentDirectoryControlHandler)
	http.HandleFunc("/ConnectionManager/scpd.xml", serveXMLDocHandler("tmpl/ConnectionManager1.xml", nil))
	http.HandleFunc("/ConnectionManager/control.xml", serviceContentDirectoryControlHandler)

	return http.Serve(srv.listener, nil)
}

func NewServer() Server {
	return Server{}
}
