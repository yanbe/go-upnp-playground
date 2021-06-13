package contentdirectory

import (
	"context"
	"encoding/xml"
	"fmt"
	"go-upnp-playground/epgstation"
	"log"
	"time"
)

var serviceURLBase string
var videoFileIdDurationMap map[epgstation.VideoFileId]time.Duration

func Setup(ServiceURLBase string) {
	log.Println("Setup ContentDirectory start")
	serviceURLBase = ServiceURLBase

	rootContainer := NewContainer("0", nil, "Root")
	recordedContainer := setupRecordedContainer(rootContainer)
	setupGenresContainer(rootContainer)
	setupChannelsContainer(rootContainer)
	setupRulesContainer(rootContainer)

	log.Printf("Setup ContentDirectory complete. %d items found", recordedContainer.ChildCount)
}

func setupRecordedContainer(parent *Container) *Container {
	recordedContainer := NewContainer("01", parent, "録画済み")
	res, err := epgstation.EPGStation.GetRecordedWithResponse(context.Background(), &epgstation.GetRecordedParams{
		IsHalfWidth: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	videoFileIdDurationMap = make(map[epgstation.VideoFileId]time.Duration)
	for _, recordedItem := range res.JSON200.Records {
		for _, videoFile := range *recordedItem.VideoFiles {
			res, err := epgstation.EPGStation.GetVideosVideoFileIdDurationWithResponse(context.Background(), epgstation.PathVideoFileId(videoFile.Id))
			if err != nil {
				log.Fatal(err)
			}
			videoFileIdDurationMap[videoFile.Id] = time.Duration(res.JSON200.Duration * float32(time.Second))
		}
	}
	for _, recordedItem := range res.JSON200.Records {
		NewItem(recordedContainer, &recordedItem, videoFileIdDurationMap)
	}
	return recordedContainer
}

func setupGenresContainer(parent *Container) *Container {
	genresContainer := NewContainer("02", parent, "ジャンル別")
	/*
		res, err := epgstation.EPGStation.GetRecordedOptionsWithResponse(context.Background())
		if err != nil {
			log.Fatal(err)
		}
			for _, genre := range res.JSON200.Genres {
				// FIXME: genre is not populated due to deserialise issue
				genreContainer := NewContainer(ObjectID(fmt.Sprintf("02%d", int(*genre.ChannelId))), genresContainer, strconv.Itoa(int(*genre.ChannelId)))
				res, err := epgstation.EPGStation.GetRecordedWithResponse(context.Background(), &epgstation.GetRecordedParams{
					IsHalfWidth: false,
					Genre:       (*epgstation.QueryProgramGenre)(genre.ChannelId),
				})
				if err != nil {
					log.Fatal(err)
				}
				for _, recordedItem := range res.JSON200.Records {
					NewItem(genreContainer, &recordedItem, videoFileIdDurationMap)
				}
			}
	*/
	return genresContainer
}

func setupChannelsContainer(parent *Container) *Container {
	channelsContainer := NewContainer("03", parent, "チャンネル別")
	resChannelInfo, err := epgstation.EPGStation.GetChannelsWithResponse(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	channelIdChannelItemMap := make(map[epgstation.ChannelId]epgstation.ChannelItem)
	for _, channelItem := range *resChannelInfo.JSON200 {
		channelIdChannelItemMap[channelItem.Id] = channelItem
	}

	res, err := epgstation.EPGStation.GetRecordedOptionsWithResponse(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, channel := range res.JSON200.Channels {
		channelName := channelIdChannelItemMap[channel.ChannelId].HalfWidthName
		channelContainer := NewContainer(ObjectID(fmt.Sprintf("03%d", int(channel.ChannelId))), channelsContainer, channelName)
		queryChannelId := epgstation.QueryChannelId(channel.ChannelId)
		res, err := epgstation.EPGStation.GetRecordedWithResponse(context.Background(), &epgstation.GetRecordedParams{
			IsHalfWidth: false,
			ChannelId:   &queryChannelId,
		})
		if err != nil {
			log.Fatal(err)
		}
		for _, recordedItem := range res.JSON200.Records {
			NewItem(channelContainer, &recordedItem, videoFileIdDurationMap)
		}
	}
	return channelsContainer
}

func setupRulesContainer(parent *Container) *Container {
	rulesContainer := NewContainer("04", parent, "ルール別")
	resRulesInfo, err := epgstation.EPGStation.GetRulesKeywordWithResponse(context.Background(), &epgstation.GetRulesKeywordParams{})
	if err != nil {
		log.Fatal(err)
	}
	for _, ruleItem := range resRulesInfo.JSON200.Items {
		ruleContainer := NewContainer(ObjectID(fmt.Sprintf("04%d", int(ruleItem.Id))), rulesContainer, ruleItem.Keyword)
		queryRuleId := epgstation.QueryRuleId(ruleItem.Id)
		res, err := epgstation.EPGStation.GetRecordedWithResponse(context.Background(), &epgstation.GetRecordedParams{
			IsHalfWidth: false,
			RuleId:      &queryRuleId,
		})
		if err != nil {
			log.Fatal(err)
		}
		for _, recordedItem := range res.JSON200.Records {
			NewItem(ruleContainer, &recordedItem, videoFileIdDurationMap)
		}
	}
	return rulesContainer
}

func GetRecordedTotal() int {
	res, err := epgstation.EPGStation.GetRecordedWithResponse(context.Background(), &epgstation.GetRecordedParams{
		IsHalfWidth: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	return res.JSON200.Total
}

func MarshalMetadata(objectID string) string {
	object := registory[ObjectID(objectID)]
	wrapper := DIDLLite{}
	wrapper.Objects = append(wrapper.Objects, &object)
	data, err := xml.Marshal(wrapper)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}

func MarshalDirectChildren(objectID string, StartingIndex int, RequestedCount int) string {
	object := registory[ObjectID(objectID)]
	container, ok := object.(*Container)
	if !ok {
		log.Fatalf("passed objectID %s not found as a container", objectID)
	}
	wrapper := DIDLLite{}
	var min, max int
	if StartingIndex < cap(container.Children) {
		min = StartingIndex
	} else {
		min = cap(container.Children)
	}
	if StartingIndex+RequestedCount <= cap(container.Children) {
		max = StartingIndex + RequestedCount
	} else {
		max = cap(container.Children)
	}
	wrapper.Objects = container.Children[min:max]
	data, err := xml.Marshal(wrapper)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}

func GetObject(objectID string) interface{} {
	return registory[ObjectID(objectID)]
}
