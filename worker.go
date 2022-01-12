package main

import (
	"daslink/dao"
	"sync"
	"time"
)

var pollingAccountList = &[]string{}

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
						removeFromPollingList(account)
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
					polling := false
					for _, a := range *pollingAccountList {
						if account == a {
							polling = true
							break
						}
					}
					if !polling {
						pollingChan <- account
						*pollingAccountList = append(*pollingAccountList, account)
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

func removeFromPollingList(account string) {
	for i, a := range *pollingAccountList {
		if a == account {
			*pollingAccountList = append((*pollingAccountList)[:i], (*pollingAccountList)[i+1:]...)
			break
		}
	}
}
