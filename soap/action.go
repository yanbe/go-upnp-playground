package soap

import (
	"context"
	"go-upnp-playground/bufferpool"
	"go-upnp-playground/epgstation"
	"html/template"
	"log"
	"net"
	"path/filepath"
	"strconv"
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

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

var funcMap = template.FuncMap{
	"date": func(t epgstation.UnixtimeMS) string { return time.Unix(int64(t)/1000, 0).In(jst).Format("2006-01-02") },
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
		var recordedItem *epgstation.RecordedItem
		var total int
		if ObjectID == "01" {
			res, err := epgstation.EPGStation.GetRecordedWithResponse(context.Background(), &epgstation.GetRecordedParams{
				IsHalfWidth: false,
			})
			if err != nil {
				log.Fatalf("could not get records from EPGStation: %s", err)
			}
			total = res.JSON200.Total
		} else {
			recordedId, err := strconv.ParseInt(ObjectID, 10, 8)
			if err != nil {
				log.Fatalf("could not parse ObjectID as int: %s", ObjectID)
			}
			res, _ := epgstation.EPGStation.GetRecordedRecordedIdWithResponse(context.Background(), epgstation.PathRecordedId(recordedId), &epgstation.GetRecordedRecordedIdParams{
				IsHalfWidth: false,
			})
			recordedItem = res.JSON200
		}
		err := template.Must(template.New("browse-metadata.xml").Funcs(funcMap).ParseFiles("tmpl/browse-metadata.xml")).
			Execute(buf, map[string]interface{}{
				"ObjectID":     ObjectID,
				"RecordedItem": recordedItem, // available when ObjectID is neither "0" nor "01"
				"Total":        total,        // available when ObjectID is "01"
				"server":       a.serverAddr,
				"filter":       parseBrowseFilter(Filter),
			})
		if err != nil {
			log.Fatal(err)
		}
		// Result, NumberReturned, TotalMatches, UpdateID
		return buf.String(), 1, 1, a.GetSystemUpdateID()
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
