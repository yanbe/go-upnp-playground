package soap

import (
	"go-upnp-playground/bufferpool"
	"go-upnp-playground/epgstation"
	"go-upnp-playground/service/contentdirectory"
	"html/template"
	"log"
	"net"
	"path/filepath"
	"strings"
	"time"
)

// 	return "dc:title,dc:creator,dc:date,upnp:class,res@size"

type browseFilter struct {
	Creator     bool
	Date        bool
	WriteStatus bool
	Res         struct {
		Size      bool
		ImportUri bool
	}
}

func parseBrowseFilter(Filter string) browseFilter {
	var filter browseFilter
	for _, field := range strings.Split(Filter, ",") {
		switch field {
		case "*":
			filter.Creator = true
			filter.Date = true
			filter.Res.Size = true
			filter.Res.ImportUri = true
			filter.WriteStatus = true
		case "dc:creator":
			filter.Creator = true
		case "dc:date":
			filter.Date = true
		case "upnp:writeStatus":
			filter.WriteStatus = true
		case "res@size":
			filter.Res.Size = true
		case "res@importUri":
			filter.Res.ImportUri = true
		}
	}
	return filter
}

type Action struct {
	serverAddr net.TCPAddr
}

var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

var funcMap = template.FuncMap{
	"date": func(t epgstation.UnixtimeMS) string { return time.Unix(int64(t)/1000, 0).In(JST).Format("2006-01-02") },
	"mimetype": func(filename string) string {
		switch filepath.Ext(filename) {
		case ".m2ts":
			return "video/mp2t"
		case ".mp4":
			return "video/mp4"
		default:
			return "application/octet-stream"
		}
	},
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
