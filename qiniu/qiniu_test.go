package qiniu_test

import (
	"testing"

	"github.com/jinzhu/configor"
	"github.com/smart-unicom/oss/qiniu"
	"github.com/smart-unicom/oss/tests"
)

type Config struct {
	AccessId  string
	AccessKey string
	Region    string
	Bucket    string
	Endpoint  string
}

type AppConfig struct {
	Private Config
	Public  Config
}

var client *qiniu.Client
var privateClient *qiniu.Client

func init() {
	config := AppConfig{}
	configor.New(&configor.Config{ENVPrefix: "QINIU"}).Load(&config)
	if len(config.Private.AccessId) == 0 {
		return
	}

	var err error
	client, err = qiniu.New(&qiniu.Config{
		AccessId:  config.Public.AccessId,
		AccessKey: config.Public.AccessKey,
		Region:    config.Public.Region,
		Bucket:    config.Public.Bucket,
		Endpoint:  config.Public.Endpoint,
	})
	if err != nil {
		panic(err)
	}

	privateClient, err = qiniu.New(&qiniu.Config{
		AccessId:   config.Private.AccessId,
		AccessKey:  config.Private.AccessKey,
		Region:     config.Private.Region,
		Bucket:     config.Private.Bucket,
		Endpoint:   config.Private.Endpoint,
		PrivateURL: true,
	})
	if err != nil {
		panic(err)
	}
}

func TestAll(t *testing.T) {
	if client == nil {
		t.Skip(`skip because of no config:


			`)
	}
	clis := []*qiniu.Client{client, privateClient}
	for _, cli := range clis {
		tests.TestAll(cli, t)
	}
}
