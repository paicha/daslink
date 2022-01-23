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
					ipfsRecordList, _ := db.FindIpfsRecordInfoByAccount(account)
					if len(ipfsRecordList) == 0 {
						dnsData.deleteDNSRecordByAccount(account)
						pollingAccounts.Delete(account)
					} else {
						priorityRecord := findPriorityRecord(ipfsRecordList[0], ipfsRecordList)
						ipfsRecord, err := dnsData.updateDNSRecord(priorityRecord)
						if err != nil {
							log.Errorf("update %s DNS record failed: %s", priorityRecord.Account, err)
						}
						go func() {
							time.Sleep(60 * time.Second)
							pollingChan <- ipfsRecord.Account
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
