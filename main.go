package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type DataInfo struct {
	MediaUId        string `json:"media_uid"`
	FileId          string `json:"file_id"`
	FileName        string `json:"file_name"`
	FilePath        string `json:"file_path"`
	ShareUrl        string `json:"share_url"`
	SharePwd        string `json:"share_pwd"`
	ShareExpiration string `json:"share_expiration"`
	ShareStatus     string `json:"share_status"`
}

var (
	aliFileDepth    = flag.Int("share_file_depth", 3, "")
	aliRefreshToken = flag.String("refresh_token", "", "")
	notionToken     = flag.String("notion_token", "", "")
	mediaDBId       = flag.String("media_db_id", "", "")
	filterFile      = flag.String("filter_file", "", "")
)

func main() {
	now := time.Now()
	flag.Parse()
	initALiYunClient(*aliRefreshToken)
	initNotionClient(*notionToken)

	split := strings.Split(*filterFile, ",")
	filterFileMap := map[string]struct{}{}
	for _, s := range split {
		filterFileMap[s] = struct{}{}
	}

	syncAliData(*aliFileDepth, filterFileMap)
	fmt.Printf("耗时：%f \n", time.Since(now).Seconds())
}
