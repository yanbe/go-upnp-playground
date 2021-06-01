package soap

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"regexp"
)

const actionNameRegexp = `"urn:schemas-upnp-org:service:ContentDirectory:1#(.+)"`

func HandleAction(r *http.Request) []byte {
	actionName := regexp.MustCompile(actionNameRegexp).FindStringSubmatch(r.Header.Get("SoapAction"))[1]
	log.Print("Action: ", actionName)

	data, _ := ioutil.ReadAll(r.Body)
	var soapReq Request
	xml.Unmarshal(data, &soapReq)
	log.Println("[device] Body:")
	reqStruct := reflect.ValueOf(soapReq.Body).FieldByName(actionName).Elem()
	argv := make([]reflect.Value, reqStruct.NumField()-1)
	for i := range argv {
		argv[i] = reqStruct.Field(i + 1) // skip XMLName field
	}
	action := Action{}
	result := reflect.ValueOf(action).MethodByName(actionName).Call(argv)
	log.Print("argv", argv, "result", result)

	var soapRes Response
	log.Print("soapRes", soapRes)
	resStructFieldPtr := reflect.ValueOf(soapRes.Body).FieldByName(actionName + "Response") // pointer to nil
	resStructPtr := reflect.New(resStructFieldPtr.Type().Elem())
	for i, v := range result {
		resStructPtr.Elem().Field(i + 1).Set(v) // skip XMLName field
	}
	/*
		soapRes.Body.BrowseResponse.Result = result[0].Interface().(string)
		soapRes.Body.BrowseResponse.TotalMatches = result[1].Interface().(int)
		soapRes.Body.BrowseResponse.NumberReturned = result[2].Interface().(int)
		soapRes.Body.BrowseResponse.UpdateID = result[3].Interface().(int)
	*/
	log.Print("resStructPtr", resStructPtr)

	reflect.ValueOf(resStructFieldPtr.Interface()).Elem().Set(resStructPtr)

	log.Print("soapRes", soapRes)
	res, _ := xml.Marshal(soapRes)
	return res
}
