package aliyun_test

import (
	"testing"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/jinzhu/configor"
	"github.com/smart-unicom/oss/aliyun"
	"github.com/smart-unicom/oss/tests"
)

type Config struct {
	AccessId  string
	AccessKey string
	Bucket    string
	Endpoint  string
}

type AppConfig struct {
	Private Config
	Public  Config
}

var client, privateClient *aliyun.Client

func init() {
	config := AppConfig{}
	err := configor.New(&configor.Config{ENVPrefix: "ALIYUN"}).Load(&config)
	if err != nil {
		panic(err)
	}

	if config.Private.AccessId == "" {
		panic("No aliyun configuration")
	}

	client = aliyun.New(&aliyun.Config{
		AccessId:  config.Public.AccessId,
		AccessKey: config.Public.AccessKey,
		Bucket:    config.Public.Bucket,
		Endpoint:  config.Public.Endpoint,
	})

	privateClient = aliyun.New(&aliyun.Config{
		AccessId:  config.Private.AccessId,
		AccessKey: config.Private.AccessKey,
		Bucket:    config.Private.Bucket,
		ACL:       aliyunoss.ACLPrivate,
		Endpoint:  config.Private.Endpoint,
	})
}

func TestAll(t *testing.T) {
	if client == nil {
		t.Skip(`skip because of no config: `)
	}

	clients := []*aliyun.Client{client, privateClient}
	for _, cli := range clients {
		tests.TestAll(cli, t)
	}
}
