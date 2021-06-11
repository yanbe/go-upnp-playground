package soap

import (
	"go-upnp-playground/bufferpool"
	"go-upnp-playground/service/contentdirectory"
	"log"
)

type Action struct {
}

func (a Action) Browse(ObjectID string, BrowseFlag string, Filter string, StartingIndex int, RequestedCount int, SortCriteria string) (string, int, int, int) {
	buf := bufferpool.NewBytesBuffer()
	defer bufferpool.PutBytesBuffer(buf)
	switch BrowseFlag {
	case "BrowseMetadata":
		return contentdirectory.MarshalMetadata(ObjectID), 1, 1, a.GetSystemUpdateID()
	case "BrowseDirectChildren":
		container := contentdirectory.GetObject(ObjectID).(*contentdirectory.Container)
		return contentdirectory.MarshalDirectChildren(ObjectID, StartingIndex, RequestedCount), container.ChildCount - StartingIndex, container.ChildCount, a.GetSystemUpdateID()
	default:
		log.Printf("invalid BrowseFlag: %s", BrowseFlag)
		// Result, NumberReturned, TotalMatches, UpdateID
		return "", 0, 0, a.GetSystemUpdateID()
	}
}

func (a Action) GetSystemUpdateID() int {
	// SystemUpdateID
	return 1
}

func (a Action) GetSearchCapabilities() string {
	// SearchCapabilities
	return "dc:title,dc:creator,dc:date,upnp:class,res@size"
}

func (a Action) GetSortCapabilities() string {
	// SortCapabilities
	return "dc:date"
}
