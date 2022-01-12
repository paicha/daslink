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
				ipfsRecordList, _ := db.FindIpfsRecordInfoByMaxId(maxId)
				for _, recordInfo := range ipfsRecordList {
					log.Debug("Found new ipfs record: ", recordInfo.Account)
					jobsChan <- recordInfo.Account
				}
				if len(ipfsRecordList) > 0 {
					maxId = ipfsRecordList[len(ipfsRecordList)-1].Id
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
