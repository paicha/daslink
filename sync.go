package main

import (
	"daslink/dao"
)

var skipContentRecordIndex = make(map[int]bool)

func runSyncContentRecords(contentRecordList []dao.TableRecordsInfo, dnsData *DNSData, jobsChan chan string) {
	// batch update dns record
	validAccounts := []string{}
	for index, contentRecord := range contentRecordList {
		// skip the same records that have already been processed
		_, skip := skipContentRecordIndex[index]
		if skip {
			continue
		}

		priorityRecord := findPriorityRecord(contentRecord, contentRecordList)

		// update CNAME and TXT record
		contentRecord, err := dnsData.updateDNSRecord(priorityRecord)
		if err != nil {
			log.Errorf("updateDNSRecord error: %s [%s - %s]", err, priorityRecord.AccountId, priorityRecord.Account)
			continue
		}

		// add accounts to worker pool
		jobsChan <- contentRecord.Account
		validAccounts = append(validAccounts, contentRecord.Account)
	}

	// batch delete invalid DNS records
	dnsData.deleteAllInvalidDNSRecord(validAccounts)
}

func findPriorityRecord(contentRecord dao.TableRecordsInfo, contentRecordList []dao.TableRecordsInfo) dao.TableRecordsInfo {
	priorityRecord := contentRecord
	count := 0
	for index, candidateRecord := range contentRecordList {
		if candidateRecord.Account == priorityRecord.Account {
			count += 1
			if count > 1 {
				skipContentRecordIndex[index] = true
				if candidateRecord.Key == priorityRecord.Key {
					// select the first record if the key is the same
					if priorityRecord.Id > candidateRecord.Id {
						priorityRecord = candidateRecord
					}
				} else if priorityRecord.Key == "skynet" {
					// Skynet record lowest priority
					priorityRecord = candidateRecord
				} else if candidateRecord.Key == "ipns" {
					// ipns record highest priority
					priorityRecord = candidateRecord
				}
			}
		}
	}
	return priorityRecord
}
