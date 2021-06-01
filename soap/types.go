package soap

import "encoding/xml"

/*
<?xml version="1.0" encoding="utf-8"?>
<s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
   <s:Body>
      <u:Browse xmlns:u="urn:schemas-upnp-org:service:ContentDirectory:1">
         <ObjectID>-1</ObjectID>
         <BrowseFlag>BrowseMetadata</BrowseFlag>
         <Filter />
         <StartingIndex>0</StartingIndex>
         <RequestedCount>0</RequestedCount>
         <SortCriteria />
      </u:Browse>
   </s:Body>
</s:Envelope>
*/

type Request struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    struct {
		XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
		Browse  *BrowseAction
	}
}

type BrowseAction struct {
	XMLName        xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 Browse"`
	ObjectID       string
	BrowseFlag     string
	Filter         string
	StartingIndex  int
	RequestedCount int
	SortCriteria   string
}

/*
<?xml version="1.0"?>
<s:Envelope
 xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"
 s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
	<s:Body>
		<u:actionNameResponse xmlns:u="urn:schemas-upnp-org:service:serviceType:v">
		<argumentName>out arg value</argumentName>
		</u:actionNameResponse>
	</s:Body>
</s:Envelope>
*/

type Response struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    struct {
		XMLName        xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
		BrowseResponse *BrowseResponse
	}
}

type BrowseResponse struct {
	XMLName        xml.Name `xml:"urn:schemas-upnp-org:service:ContentDirectory:1 BrowseResponse"`
	Result         string
	NumberReturned int
	TotalMatches   int
	UpdateID       int
}
