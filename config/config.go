package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/scorpiotzh/mylog"
	"github.com/scorpiotzh/toolib"
)

var (
	Cfg CfgServer
	log = mylog.NewLogger("config", mylog.LevelDebug)
)

func InitCfg(configFilePath string) error {
	if configFilePath == "" {
		configFilePath = "./config/config.yaml"
	}
	log.Info("read from config: ", configFilePath)
	if err := toolib.UnmarshalYamlFile(configFilePath, &Cfg); err != nil {
		return fmt.Errorf("UnmarshalYamlFile err:%s", err.Error())
	}
	log.Info("config file: ", toolib.JsonString(Cfg))
	return nil
}

func AddCfgFileWatcher(configFilePath string) (*fsnotify.Watcher, error) {
	if configFilePath == "" {
		configFilePath = "./config/config.yaml"
	}
	return toolib.AddFileWatcher(configFilePath, func() {
		log.Info("update config file: ", configFilePath)
		if err := toolib.UnmarshalYamlFile(configFilePath, &Cfg); err != nil {
			log.Error("UnmarshalYamlFile err:", err.Error())
		}
		log.Info("new config file: ", toolib.JsonString(Cfg))
	})
}

type CfgServer struct {
	DB struct {
		Mysql DbMysql `json:"mysql" yaml:"mysql"`
	} `json:"db" yaml:"db"`
	CloudFlare struct {
		ApiKey   string `json:"api_key" yaml:"api_key"`
		ApiEmail string `json:"api_email" yaml:"api_email"`
		ZoneName string `json:"zone_name" yaml:"zone_name"`
	} `json:"cloudflare" yaml:"cloudflare"`
	IPFS struct {
		Gateway string `json:"gateway" yaml:"gateway"`
	} `json:"ipfs" yaml:"ipfs"`
	HostName struct {
		Suffix string `json:"suffix" yaml:"suffix"`
	} `json:"hostname" yaml:"hostname"`
}

type DbMysql struct {
	LogMode     bool   `json:"log_mode" yaml:"log_mode"`
	Addr        string `json:"addr" yaml:"addr"`
	User        string `json:"user" yaml:"user"`
	Password    string `json:"password" yaml:"password"`
	DbName      string `json:"db_name" yaml:"db_name"`
	MaxOpenConn int    `json:"max_open_conn" yaml:"max_open_conn"`
	MaxIdleConn int    `json:"max_idle_conn" yaml:"max_idle_conn"`
}
