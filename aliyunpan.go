package main

import (
	"context"
	"fmt"
	"github.com/jomei/notionapi"
	jsoniter "github.com/json-iterator/go"
	"github.com/tickstep/aliyunpan-api/aliyunpan"
	"log"
)

var (
	aliYunClient *aliyunpan.PanClient
	ui           *aliyunpan.UserInfo
)

func initALiYunClient(refreshToken string) {
	if len(refreshToken) == 0 {
		fmt.Println("ALI_REFRESH_TOKEN empty")
	}
	webToken, err := aliyunpan.GetAccessTokenFromRefreshToken(refreshToken)
	if err != nil {
		fmt.Println("get acccess token error")
		return
	}

	// pan client
	aliYunClient = aliyunpan.NewPanClient(*webToken, aliyunpan.AppLoginToken{})
	getUserInfo()
}

func objToJsonStr(v interface{}) string {
	r, _ := jsoniter.MarshalToString(v)
	return r
}

func getUserInfo() {
	tui, err := aliYunClient.GetUserInfo()
	if err != nil {
		log.Fatal(err)
	}
	ui = tui
}

func syncAliData(depth int, filterFile map[string]struct{}) {
	list, apiError := aliYunClient.ShareLinkList(ui.UserId)
	if apiError != nil {
		fmt.Println(apiError)
	}
	fileId2ShareInfoMap := map[string]*aliyunpan.ShareEntity{}
	for _, entity := range list {
		for _, fId := range entity.FileIdList {
			fileId2ShareInfoMap[fId] = entity
		}
	}
	fl1 := aliYunClient.FilesDirectoriesRecurseListDepth(ui.FileDriveId, "/", depth, filterFile, nil)
	fmt.Printf("file num: %d \n", len(fl1))
	infos := make([]*DataInfo, 0, len(fl1))
	for _, entity := range fl1 {
		s, ok := fileId2ShareInfoMap[entity.FileId]
		if !ok {
			s, _ = aliYunClient.ShareLinkCreate(aliyunpan.ShareCreateParam{
				DriveId:    ui.FileDriveId,
				SharePwd:   "8888",
				FileIdList: []string{entity.FileId},
			})
		}
		if s == nil {
			continue
		}
		infos = append(infos, &DataInfo{
			FileId:          entity.FileId,
			FileName:        entity.FileName,
			FilePath:        entity.Path,
			ShareUrl:        s.ShareUrl,
			SharePwd:        s.SharePwd,
			ShareExpiration: s.Expiration,
			ShareStatus:     s.Status,
		})
	}
	fmt.Printf("shareFile num: %d \n", len(infos))

	sendDataToNotion(context.TODO(), infos, notionapi.DatabaseID(*mediaDBId))
}
