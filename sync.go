package main

import (
	"daslink/dao"
)

var skipIpfsRecordIndex = make(map[int]bool)

func runSyncIpfsRecords(ipfsRecordList []dao.TableRecordsInfo, dnsData *DNSData, jobsChan chan string) {
	// batch update dns record
	validAccounts := []string{}
	for index, ipfsRecord := range ipfsRecordList {
		// skip the same records that have already been processed
		_, skip := skipIpfsRecordIndex[index]
		if skip {
			continue
		}

		priorityRecord := findPriorityRecord(ipfsRecord, ipfsRecordList)

		// update CNAME and TXT record
		ipfsRecord, err := dnsData.updateDNSRecord(priorityRecord)
		if err != nil {
			log.Errorf("updateDNSRecord error: %s", err)
			continue
		}

		// add accounts to worker pool
		jobsChan <- ipfsRecord.Account
		validAccounts = append(validAccounts, ipfsRecord.Account)
	}

	// batch delete invalid DNS records
	dnsData.deleteAllInvalidDNSRecord(validAccounts)
}

func findPriorityRecord(ipfsRecord dao.TableRecordsInfo, ipfsRecordList []dao.TableRecordsInfo) dao.TableRecordsInfo {
	priorityRecord := ipfsRecord
	count := 0
	for index, candidateRecord := range ipfsRecordList {
		if candidateRecord.Account == priorityRecord.Account {
			count += 1
			if count > 1 {
				skipIpfsRecordIndex[index] = true
				if candidateRecord.Key == priorityRecord.Key {
					// select the first record if the key is the same
					if priorityRecord.Id > candidateRecord.Id {
						priorityRecord = candidateRecord
					}
				} else if candidateRecord.Key == "ipns" {
					// ipns record priority
					priorityRecord = candidateRecord
				}
			}
		}
	}
	return priorityRecord
}
