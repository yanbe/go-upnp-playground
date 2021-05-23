package device

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/google/uuid"
)

const deviceDescritpionXml = `<?xml version="1.0"?> 
	<root xmlns="urn:schemas-upnp-org:device-1-0"> 
	<specVersion> 
		<major>1</major> 
		<minor>0</minor> 
	</specVersion> 
	<URLBase>http://%s/</URLBase>
	<device> 
		<deviceType>urn:schemas-upnp-org:device:MediaServer:1</deviceType>
		<friendlyName>go-upnp-playground</friendlyName> 
		<manufacturer>manufacturer name</manufacturer> 
		<manufacturerURL>URL to manufacturer site</manufacturerURL> 
		<modelDescription>long user-friendly title</modelDescription> 
		<modelName>go-upnp-playground</modelName> 
		<modelNumber>0.0.1</modelNumber> 
		<modelURL>https://github.com/yanbe/go-upnp-playground</modelURL> 
		<serialNumber>manufacturer's serial number</serialNumber> 
		<UDN>uuid:%s</UDN> 
		<UPC>Universal Product Code</UPC> 
		<iconList> 
			<icon> 
				<mimetype>image/png</mimetype> 
				<width>120</width> 
				<height>120</height> 
				<depth>8</depth> 
				<url>http://192.168.10.9:1737/dmrIcon_120.png</url> 
			</icon> 
		</iconList> 
		<serviceList> 
			<service>
				<serviceType>urn:schemas-upnp-org:service:ConnectionManager:1</serviceType>
				<serviceId>urn:schemas-upnp-org:service:ConnectionManager</serviceId>
				<SCPDURL>_urn-schemas-upnp-org-service-ConnectionManager_scpd.xml</SCPDURL>
				<controlURL>_urn-schemas-upnp-org-service-ConnectionManager_control</controlURL>
				<eventSubURL>_urn-schemas-upnp-org-service-ConnectionManager_event</eventSubURL>
			</service>
			<service>
				<serviceType>urn:schemas-upnp-org:service:ContentDirectory:1</serviceType>
				<serviceId>urn:schemas-upnp-org:service:ContentDirectory</serviceId>
				<SCPDURL>_urn-schemas-upnp-org-service-ContentDirectory_scpd.xml</SCPDURL>
				<controlURL>_urn-schemas-upnp-org-service-ContentDirectory_control</controlURL>
				<eventSubURL>_urn-schemas-upnp-org-service-ContentDirectory_event</eventSubURL>
			</service>		
		</serviceList> 
		<presentationURL>http://192.168.10.10:8888/</presentationURL> 
	</device> 
</root>`

// A Server defines parameters for running an HTTPU server.
type Server struct {
	uuid     uuid.UUID
	Addr     *net.TCPAddr
	listener net.Listener
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wb := bufio.NewWriter(w)
	fmt.Fprintf(wb, deviceDescritpionXml, srv.Addr, srv.uuid)
	wb.Flush()
	log.Printf("[device] Got %s %s message from %v: %v", r.Method, r.URL.Path, r.RemoteAddr, r.Header)
}

func (srv *Server) Listen(hostIP string) {
	srv.listener, _ = net.Listen("tcp", hostIP+":0")
	srv.Addr = srv.listener.Addr().(*net.TCPAddr)
	fmt.Println("Listening", srv.Addr)
}

func (srv *Server) Serve() error {
	http.Handle("/", srv) // ハンドラを登録してウェブページを表示させる
	return http.Serve(srv.listener, nil)
}

func NewServer(uuid uuid.UUID) Server {
	return Server{
		uuid: uuid,
	}
}
