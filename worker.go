package main

import (
	"daslink/dao"
	"sync"
	"time"
)

var pollingAccounts = sync.Map{}

func runWorker(wg *sync.WaitGroup, db *dao.DbDao, dnsData *DNSData, jobsChan chan string) {
	pollingChan := make(chan string, len(jobsChan)*2)
	for i := 0; i < (len(jobsChan)/50 + 1); i++ {
		wg.Add(1)
		go func() {
			for {
				select {
				default:
					account := <-pollingChan
					log.Infof("polling %s DNS record", account)
					contentRecordList, _ := db.FindContentRecordInfoByAccount(account)
					if len(contentRecordList) == 0 {
						dnsData.deleteDNSRecordByAccount(account)
						pollingAccounts.Delete(account)
					} else {
						priorityRecord := findPriorityRecord(contentRecordList[0], contentRecordList)
						contentRecord, err := dnsData.updateDNSRecord(priorityRecord)
						if err != nil {
							log.Errorf("update %s DNS record failed: %s", priorityRecord.Account, err)
						}
						go func() {
							time.Sleep(60 * time.Second)
							pollingChan <- contentRecord.Account
						}()
					}
				case account := <-jobsChan:
					_, polling := pollingAccounts.Load(account)
					if !polling {
						pollingChan <- account
						pollingAccounts.Store(account, true)
					}
				case <-ctxServer.Done():
					log.Info("worker exit")
					wg.Done()
					return
				}
			}
		}()
	}
}
