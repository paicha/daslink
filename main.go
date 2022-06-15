package main

import (
	"context"
	"daslink/config"
	"daslink/dao"
	"fmt"
	"os"
	"sync"

	"github.com/scorpiotzh/mylog"
	"github.com/scorpiotzh/toolib"
	"github.com/urfave/cli/v2"
)

var (
	log               = mylog.NewLogger("main", mylog.LevelDebug)
	exit              = make(chan struct{})
	ctxServer, cancel = context.WithCancel(context.Background())
	wgServer          = sync.WaitGroup{}
)

func main() {
	log.Debugf("server start:")
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
		},
		Action: runServer,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runServer(ctx *cli.Context) error {
	// config
	configFilePath := ctx.String("config")
	if err := config.InitCfg(configFilePath); err != nil {
		return err
	}

	// db
	cfgMysql := config.Cfg.DB.Mysql
	db, err := dao.NewGormDataBase(cfgMysql.Addr, cfgMysql.User, cfgMysql.Password, cfgMysql.DbName, cfgMysql.MaxOpenConn, cfgMysql.MaxIdleConn)
	if err != nil {
		return fmt.Errorf("NewGormDataBase err:%s", err.Error())
	}
	dbDao := dao.Initialize(db, cfgMysql.LogMode)
	log.Info("db ok")

	// dns
	cfgCloudflare := config.Cfg.CloudFlare
	dnsData, err := NewDNSData(cfgCloudflare.ApiKey, cfgCloudflare.ApiEmail, cfgCloudflare.ZoneName, config.Cfg.Gateway.Ipfs, config.Cfg.Gateway.Skynet, config.Cfg.HostName.Suffix)
	if err != nil {
		return fmt.Errorf("NewDNSData err:%s", err.Error())
	}
	log.Info("dns data ok")

	// read all das accounts that has ipfs/ipns/skynet record
	contentRecordList, _ := dbDao.FindRecordInfoByKeys([]string{"ipfs", "ipns", "skynet"})

	jobsChanLength := len(contentRecordList)
	if jobsChanLength == 0 {
		jobsChanLength = 1
	}
	jobsChan := make(chan string, jobsChanLength)

	maxId := uint64(0)
	if len(contentRecordList) > 0 {
		maxId = contentRecordList[len(contentRecordList)-1].Id
	}

	runWatcher(&wgServer, dbDao, maxId, jobsChan)
	log.Info("Watching new content records...")

	runSyncContentRecords(contentRecordList, dnsData, jobsChan)
	log.Info("All content records have been synchronized")

	runWorker(&wgServer, dbDao, dnsData, jobsChan)
	log.Info("Worker started")

	// quit monitor
	toolib.ExitMonitoring(func(sig os.Signal) {
		log.Warn("ExitMonitoring:", sig.String())
		cancel()
		log.Warn("Wait for worker to finish...")
		wgServer.Wait()
		exit <- struct{}{}
	})

	<-exit
	log.Warn("success exit server. bye bye!")
	return nil
}
