package main

import (
	"context"
	"github.com/jomei/notionapi"
	"log"
)

const (
	Name      = "Name"
	FileId    = "file_id"
	AliYunPan = "阿里云盘"
	ShareCode = "share_code"
)

var notionApiClient *notionapi.Client

func initNotionClient(token string) {
	notionApiClient = notionapi.NewClient(notionapi.Token(token))
}

const (
	DefaultBatchSize = 1
	DefaultWorker    = 10
)

func sendDataToNotion(ctx context.Context, dataList []*DataInfo, dbId notionapi.DatabaseID) {
	runner := NewBatchRunner(len(dataList), DefaultBatchSize, DefaultWorker)
	for runner.Iter() {
		tmpDataList := dataList[runner.Begin():runner.End()]
		runner.Run(func() {
			syncData(ctx, dbId, tmpDataList[0])
		})
	}
}

func syncData(ctx context.Context, dbId notionapi.DatabaseID, info *DataInfo) {
	response, err := notionApiClient.Database.Query(ctx, dbId, &notionapi.DatabaseQueryRequest{
		Filter: notionapi.PropertyFilter{
			Property: FileId,
			RichText: &notionapi.TextFilterCondition{
				Equals: info.FileId,
			},
		},
		PageSize: 1,
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(response.Results) == 0 {
		_, err := notionApiClient.Page.Create(ctx, &notionapi.PageCreateRequest{
			Parent: notionapi.Parent{
				Type:       notionapi.ParentTypeDatabaseID,
				DatabaseID: dbId,
			},
			Properties: notionapi.Properties{
				Name: notionapi.TitleProperty{
					Title: []notionapi.RichText{
						{Text: &notionapi.Text{Content: info.FileName}},
					},
				},
				FileId: notionapi.RichTextProperty{
					Type: notionapi.PropertyTypeRichText,
					RichText: []notionapi.RichText{
						{
							Text: &notionapi.Text{Content: info.FileId},
						},
					},
				},
				AliYunPan: notionapi.URLProperty{
					Type: notionapi.PropertyTypeURL,
					URL:  info.ShareUrl,
				},
				ShareCode: notionapi.RichTextProperty{
					Type: notionapi.PropertyTypeRichText,
					RichText: []notionapi.RichText{
						{
							Text: &notionapi.Text{Content: info.SharePwd},
						},
					},
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		var urlChange, nameChange bool
		page := response.Results[0]
		urlProperty, ok := page.Properties[AliYunPan].(*notionapi.URLProperty)
		if !(ok && urlProperty.URL == info.ShareUrl) {
			urlChange = true
		}
		nameProperty, ok := page.Properties[Name].(*notionapi.TitleProperty)
		if !(ok && len(nameProperty.Title) != 0 && nameProperty.Title[0].Text.Content == info.FileName) {
			nameChange = true
		}
		if urlChange || nameChange {
			_, err := notionApiClient.Page.Update(ctx, notionapi.PageID(page.ID), &notionapi.PageUpdateRequest{
				Properties: notionapi.Properties{
					Name: notionapi.TitleProperty{
						Title: []notionapi.RichText{
							{Text: &notionapi.Text{Content: info.FileName}},
						},
					},
					AliYunPan: notionapi.URLProperty{
						Type: notionapi.PropertyTypeURL,
						URL:  info.ShareUrl,
					},
					ShareCode: notionapi.RichTextProperty{
						Type: notionapi.PropertyTypeRichText,
						RichText: []notionapi.RichText{
							{
								Text: &notionapi.Text{Content: info.SharePwd},
							},
						},
					},
				},
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
