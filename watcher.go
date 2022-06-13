package main

import (
	"daslink/dao"
	"sync"
	"time"
)

func runWatcher(wg *sync.WaitGroup, db *dao.DbDao, maxId uint64, jobsChan chan string) {
	wg.Add(1)
	go func() {
		for {
			select {
			default:
				contentRecordList, _ := db.FindContentRecordInfoByMaxId(maxId)
				for _, recordInfo := range contentRecordList {
					log.Debug("Found new content record: ", recordInfo.Account)
					jobsChan <- recordInfo.Account
				}
				if len(contentRecordList) > 0 {
					maxId = contentRecordList[len(contentRecordList)-1].Id
				}
				time.Sleep(60 * time.Second)
			case <-ctxServer.Done():
				log.Info("watch exit")
				wg.Done()
				return
			}
		}
	}()
}
