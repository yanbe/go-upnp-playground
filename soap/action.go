package soap

import (
	"context"
	"encoding/xml"
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
		object := contentdirectory.Registory[contentdirectory.ObjectID(ObjectID)]
		container, ok := object.(*contentdirectory.Container)
		wrapper := DIDLLite{}
		if ok {
			wrapper.Containers = append(wrapper.Containers, container)
			data, err := xml.Marshal(wrapper)
			if err != nil {
				log.Fatal(err)
			}
			return string(data), 1, 1, a.GetSystemUpdateID()
		} else {
			item := object.(*contentdirectory.Item)
			wrapper.Items = append(wrapper.Items, item)
			data, err := xml.Marshal(item)
			if err != nil {
				log.Fatal(err)
			}
			return string(data), 1, 1, a.GetSystemUpdateID()
		}
	case "BrowseDirectChildren":
		offset := epgstation.Offset(StartingIndex)
		params := epgstation.GetRecordedParams{
			IsHalfWidth: false,
			Offset:      &offset,
		}
		if RequestedCount > 0 { // to avoid EPGStation API error, pass limit parameter to api endpoint only if it makes sense
			limit := epgstation.Limit(RequestedCount)
			params.Limit = &limit
		}
		if SortCriteria == "+dc:date" {
			isReverse := epgstation.IsReverse(true)
			params.IsReverse = &isReverse
		}
		res, err := epgstation.EPGStation.GetRecordedWithResponse(context.Background(), &params)
		if err != nil {
			log.Fatalf("epgstation client getrecorded error: %s", err)
		}
		err = template.Must(template.New("browse-children.xml").Funcs(funcMap).ParseFiles("tmpl/browse-children.xml")).
			Execute(buf, map[string]interface{}{
				"ObjectID":       ObjectID,
				"Records":        res.JSON200.Records,
				"Total":          res.JSON200.Total, // used when ObjectID is "01"
				"StartingIndex":  StartingIndex,     // used when ObjectID is "0"
				"RequestedCount": RequestedCount,    // used when ObjectID is "0"
				"server":         a.serverAddr,
				"filter":         parseBrowseFilter(Filter),
			})
		if err != nil {
			log.Fatal(err)
		}
		switch ObjectID {
		case "0":
			// Result, NumberReturned, TotalMatches, UpdateID
			if StartingIndex == 0 {
				return buf.String(), 1, 1, a.GetSystemUpdateID()
			} else {
				return buf.String(), 0, 1, a.GetSystemUpdateID()
			}
		case "01":
			return buf.String(), len(res.JSON200.Records), res.JSON200.Total, a.GetSystemUpdateID()
		default:
			return buf.String(), 0, 0, a.GetSystemUpdateID()
		}

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
