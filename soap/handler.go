package soap

import (
	"encoding/xml"
	"fmt"
	"go-upnp-playground/epgstation"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reflect"
	"regexp"
)

const actionNameRegexp = `"urn:schemas-upnp-org:service:ContentDirectory:1#(.+)"`

var soapAction = Action{}

func SetupClient(epgstationAddr net.TCPAddr) {
	soapAction.serverAddr = epgstationAddr
	var err error
	soapAction.epgStationClient, err = epgstation.NewClientWithResponses(fmt.Sprintf("http://%s:%d/api", epgstationAddr.IP, epgstationAddr.Port))
	if err != nil {
		log.Fatalf("epgstation client init error: %s", err)
	}

}

func HandleAction(r *http.Request) []byte {
	actionName := regexp.MustCompile(actionNameRegexp).FindStringSubmatch(r.Header.Get("SoapAction"))[1]

	data, _ := ioutil.ReadAll(r.Body)
	var soapReq Request
	xml.Unmarshal(data, &soapReq)
	reqStruct := reflect.ValueOf(soapReq.Body).FieldByName(actionName).Elem()
	argv := make([]reflect.Value, reqStruct.NumField()-1)
	for i := range argv {
		argv[i] = reqStruct.Field(i + 1) // skip XMLName field
	}
	result := reflect.ValueOf(&soapAction).MethodByName(actionName).Call(argv)

	var soapRes Response
	soapRes.EncodingStyle = "http://schemas.xmlsoap.org/soap/encoding/"
	resStructFieldPtr := reflect.ValueOf(soapRes.Body).FieldByName(actionName + "Response") // pointer to nil
	resStructPtr := reflect.New(resStructFieldPtr.Type().Elem())
	for i, v := range result {
		resStructPtr.Elem().Field(i + 1).Set(v) // skip XMLName field
	}
	reflect.ValueOf(&soapRes.Body).Elem().FieldByName(actionName + "Response").Set(resStructPtr)
	res, _ := xml.Marshal(soapRes)
	return res
}
