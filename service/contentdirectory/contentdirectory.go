package contentdirectory

import (
	"context"
	"encoding/xml"
	"fmt"
	"go-upnp-playground/epgstation"
	"log"
	"path/filepath"
	"strconv"
	"time"
)

var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

type contentDirectory struct {
	Root *Container
}

type ObjectID string

var ContentDirectory contentDirectory
var Registory = make(map[ObjectID]interface{})

func Setup() {
	log.Println("Setup ContentDirectory start")
	rootContainer := NewContainer("0", nil, "Root")
	ContentDirectory = contentDirectory{
		Root: rootContainer,
	}
	recordedContainer := NewContainer("01", ContentDirectory.Root, "Recorded")

	resChannels, err := epgstation.EPGStation.GetChannelsWithResponse(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	channelMap := make(map[epgstation.ChannelId]*epgstation.ChannelItem)
	for _, channelItem := range *resChannels.JSON200 {
		channelMap[channelItem.Id] = &channelItem
	}

	res, err := epgstation.EPGStation.GetRecordedWithResponse(context.Background(), &epgstation.GetRecordedParams{
		IsHalfWidth: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, recordedItem := range res.JSON200.Records {
		NewItem(recordedContainer, &recordedItem, channelMap[*recordedItem.ChannelId])
	}
	log.Printf("Setup ContentDirectory complete. %d items found", recordedContainer.ChildCount)
}

type Container struct {
	XMLName xml.Name `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ container"`

	Id         ObjectID `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ id,attr"`
	ParentID   ObjectID `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ parentID,attr"`
	Title      string   `xml:"http://purl.org/dc/elements/1.1/ title"`
	Class      string   `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ class"`
	Restricted string   `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ restricted,attr"`

	ChildCount int           `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ childCount,attr"`
	Children   []interface{} `xml:"-"`
}

func (c *Container) AppendChild(child *interface{}) {
	c.Children = append(c.Children, child)
}

type Item struct {
	XMLName xml.Name `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ item"`

	Id         ObjectID `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ id,attr"`
	ParentID   ObjectID `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ parentID,attr"`
	Title      string   `xml:"http://purl.org/dc/elements/1.1/ title"`
	Class      string   `xml:"urn:schemas-upnp-org:metadata-1-0/upnp/ class"`
	Restricted string   `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ restricted,attr"`

	Creator   string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Date      string `xml:"http://purl.org/dc/elements/1.1/ date"`
	Resources *[]Res
}

type Res struct {
	XMLName      xml.Name `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ res"`
	ProtocolInfo string   `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ protocolInfo,attr"`
	URL          string   `xml:"urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/ res,chardata"`
}

func NewContainer(Id ObjectID, Parent *Container, Title string) *Container {
	var objectID, parentID ObjectID
	switch Parent {
	case nil:
		parentID, objectID = ObjectID("-1"), ObjectID("0")
	default:
		parentID, objectID = Parent.Id, Id
	}
	container := Container{
		Id:         objectID,
		ParentID:   parentID,
		Title:      Title,
		Class:      "object.container",
		Restricted: "true",
		Children:   make([]interface{}, 0),
	}
	Registory[container.Id] = container
	if Parent != nil {
		Parent.ChildCount += 1
		Parent.Children = append(Parent.Children, &container)
	}
	return &container
}

func mimeType(Filename string) string {
	switch filepath.Ext(Filename) {
	case ".m2ts":
		return "video/mp2t"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}

func NewItem(Parent *Container, recordedItem *epgstation.RecordedItem, channelItem *epgstation.ChannelItem) *Item {
	if Parent == nil {
		log.Fatal("container is required for item")
	}

	resources := make([]Res, len(*recordedItem.VideoFiles))
	for i, videoFile := range *recordedItem.VideoFiles {
		resources[i].ProtocolInfo = fmt.Sprintf("http-get:*:%s:*", mimeType(*videoFile.Filename))
		resources[i].URL = fmt.Sprintf("%s/videos/%d", epgstation.ServerAPIRoot, videoFile.Id)
	}
	item := Item{
		Id:         ObjectID(strconv.Itoa(int(recordedItem.Id))),
		ParentID:   Parent.Id,
		Title:      recordedItem.Name,
		Class:      "object.item.videoItem",
		Restricted: "true",

		Resources: &resources,

		Creator: channelItem.Name,
		Date:    time.Unix(int64(recordedItem.StartAt)/1000, 0).In(JST).Format("2006-01-02"),
	}
	Registory[item.Id] = item
	Parent.Children = append(Parent.Children, &item)
	Parent.ChildCount++
	return &item
}
