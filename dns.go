package main

import (
	"daslink/dao"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/miekg/dns"
	"golang.org/x/net/idna"
)

type DNSData struct {
	api            *cloudflare.API
	zoneID         string
	ipfsCname      string
	skynetCname    string
	hostNameSuffix string
	cnameRecords   *[]cloudflare.DNSRecord
	txtRecords     *[]cloudflare.DNSRecord
}

func NewDNSData(apiKey, apiEmail, zoneName, ipfsCname, skynetCname, hostNameSuffix string) (*DNSData, error) {
	api, err := cloudflare.New(apiKey, apiEmail)
	if err != nil {
		return nil, fmt.Errorf("cloudflare err:%s", err.Error())
	}
	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		return nil, fmt.Errorf("cloudflare ZoneIDByName err:%s", err.Error())
	}

	d := DNSData{
		api:            api,
		zoneID:         zoneID,
		ipfsCname:      ipfsCname,
		skynetCname:    skynetCname,
		hostNameSuffix: hostNameSuffix,
	}

	if err = d.getAllDNSRecord(); err != nil {
		return nil, fmt.Errorf("cloudflare NewDNSData err:%s", err.Error())
	}

	return &d, nil
}

func (d *DNSData) getAllDNSRecord() error {
	// get all CNAME records from cloudflare DNS
	cnameRecords, err := d.api.DNSRecords(ctxServer, d.zoneID, cloudflare.DNSRecord{Type: "CNAME"})
	if err != nil {
		return err
	}
	log.Infof("find %d CNAME records", len(cnameRecords))

	// get all TXT records from cloudflare DNS
	txtRecords, err := d.api.DNSRecords(ctxServer, d.zoneID, cloudflare.DNSRecord{Type: "TXT"})
	if err != nil {
		return err
	}
	log.Infof("find %d TXT records", len(txtRecords))

	d.cnameRecords = &cnameRecords
	d.txtRecords = &txtRecords

	return nil
}

func (d *DNSData) updateDNSRecord(contentRecord dao.TableRecordsInfo) (dao.TableRecordsInfo, error) {
	ttl, _ := strconv.Atoi(contentRecord.Ttl)
	if ttl < 60 {
		ttl = 60
	} else if ttl > 86400 {
		ttl = 86400
	}
	var isDomainName bool
	for _, recordType := range []string{"CNAME", "TXT"} {
		newRecord := cloudflare.DNSRecord{}
		oldRecords := []cloudflare.DNSRecord{}
		if recordType == "CNAME" {
			oldRecords = *d.cnameRecords
			content := d.ipfsCname
			proxied := true
			// CNAME content for skynet
			if contentRecord.Key == "skynet" {
				content = d.skynetCname
			}
			newRecord = cloudflare.DNSRecord{
				Type:    recordType,
				Name:    contentRecord.Account + d.hostNameSuffix,
				Content: content,
				TTL:     ttl,
				Proxied: &proxied,
			}
		} else if recordType == "TXT" {
			oldRecords = *d.txtRecords
			value := contentRecord.Value
			// compatible with ipfs://xxx sia://xxx https://ipfs.io/ipfs/xxx
			if contentRecord.Key == "skynet" || contentRecord.Key == "ipfs" {
				re := regexp.MustCompile(`([0-9A-Za-z-_]{46})`)
				if results := re.FindStringSubmatch(value); len(results) == 2 {
					value = results[1]
				}
			}
			txt := "dnslink=/" + contentRecord.Key + "/" + value
			if contentRecord.Key == "skynet" {
				txt = "dnslink=/skynet-ns/" + value
			}
			newRecord = cloudflare.DNSRecord{
				Type:    recordType,
				Name:    "_dnslink." + contentRecord.Account + d.hostNameSuffix,
				Content: txt,
				TTL:     ttl,
			}
		}

		recordExist := false
		oldRecordId := ""
		for _, oldRecord := range oldRecords {
			oldRecordName, _ := idna.ToUnicode(oldRecord.Name)
			if newRecord.Name == oldRecordName {
				log.Infof("%s record exist: %s", recordType, newRecord.Name)
				recordExist = true
				if newRecord.Content != oldRecord.Content || (recordType == "TXT" && newRecord.TTL != oldRecord.TTL) {
					log.Infof("%s record exist and NEED to update: %s", recordType, newRecord.Name)
					oldRecordId = oldRecord.ID
				}
				break
			}
		}

		_, isDomainName = dns.IsDomainName(newRecord.Name)
		if !isDomainName {
			return contentRecord, fmt.Errorf("%s not a vlaid domain name", newRecord.Name)
		} else if !recordExist {
			record, err := d.api.CreateDNSRecord(ctxServer, d.zoneID, newRecord)
			if err != nil {
				return contentRecord, fmt.Errorf("cloudflare CreateDNSRecord err: %s", err.Error())
				//log.Fatalf("cloudflare CreateDNSRecord err:%s", err)
			}
			if recordType == "CNAME" {
				*d.cnameRecords = append(*d.cnameRecords, record.Result)
			} else if recordType == "TXT" {
				*d.txtRecords = append(*d.txtRecords, record.Result)
			}
			log.Debugf("Successfully created %s record: %s", recordType, record.Result.Name)
		} else if oldRecordId != "" {
			err := d.api.UpdateDNSRecord(ctxServer, d.zoneID, oldRecordId, newRecord)
			if err != nil {
				return contentRecord, fmt.Errorf("cloudflare UpdateDNSRecord err: %s", err.Error())
				//log.Fatalf("cloudflare UpdateDNSRecord err:%s", err)
			}
			if recordType == "CNAME" {
				for i, cnameRecord := range *d.cnameRecords {
					if cnameRecord.ID == oldRecordId {
						(*d.cnameRecords)[i].Content = newRecord.Content
						(*d.cnameRecords)[i].TTL = newRecord.TTL
					}
				}
			} else if recordType == "TXT" {
				for i, txtRecord := range *d.txtRecords {
					if txtRecord.ID == oldRecordId {
						(*d.txtRecords)[i].Content = newRecord.Content
						(*d.txtRecords)[i].TTL = newRecord.TTL
					}
				}
			}
			log.Debugf("Successfully updated %s record: %s", recordType, newRecord.Name)
		}
	}
	return contentRecord, nil
}

func (d *DNSData) deleteAllInvalidDNSRecord(validAccounts []string) {
	for _, oldRecords := range [][]cloudflare.DNSRecord{*d.cnameRecords, *d.txtRecords} {
		for _, oldRecord := range oldRecords {
			if (oldRecord.Type == "CNAME" && (oldRecord.Content == d.ipfsCname || oldRecord.Content == d.skynetCname)) || oldRecord.Type == "TXT" && strings.HasPrefix(oldRecord.Name, "_dnslink.") {
				recordOutDated := true
				for _, validAccount := range validAccounts {
					oldRecordName, _ := idna.ToUnicode(oldRecord.Name)
					if oldRecordName == validAccount+d.hostNameSuffix || oldRecordName == "_dnslink."+validAccount+d.hostNameSuffix {
						recordOutDated = false
						break
					}
				}
				if recordOutDated {
					log.Infof("%s record NEED to delete: %s", oldRecord.Type, oldRecord.Name)
					d.deleteDNSRecord(oldRecord)
				}
			}
		}
	}
}

func (d *DNSData) deleteDNSRecord(record cloudflare.DNSRecord) {
	err := d.api.DeleteDNSRecord(ctxServer, d.zoneID, record.ID)
	if err != nil {
		log.Fatalf("cloudflare DeleteDNSRecord err:%s", err)
	} else {
		if record.Type == "CNAME" {
			for i, cnameRecord := range *d.cnameRecords {
				if cnameRecord.ID == record.ID {
					*d.cnameRecords = append((*d.cnameRecords)[:i], (*d.cnameRecords)[i+1:]...)
					break
				}
			}
		} else if record.Type == "TXT" {
			for i, txtRecord := range *d.txtRecords {
				if txtRecord.ID == record.ID {
					*d.txtRecords = append((*d.txtRecords)[:i], (*d.txtRecords)[i+1:]...)
					break
				}
			}
		}
		log.Debugf("Successfully delete %s record: %s", record.Type, record.Name)
	}
}

func (d *DNSData) deleteDNSRecordByAccount(account string) {
	for _, oldRecords := range [][]cloudflare.DNSRecord{*d.cnameRecords, *d.txtRecords} {
		for _, oldRecord := range oldRecords {
			oldRecordName, _ := idna.ToUnicode(oldRecord.Name)
			if oldRecordName == account+d.hostNameSuffix || oldRecordName == "_dnslink."+account+d.hostNameSuffix {
				log.Infof("%s record NEED to delete: %s", oldRecord.Type, oldRecord.Name)
				d.deleteDNSRecord(oldRecord)
			}
		}
	}
}
