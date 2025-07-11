package synology_test

import (
	"testing"

	"github.com/jinzhu/configor"
	"github.com/smart-unicom/oss/synology"
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

var client *synology.Client
var privateClient *synology.Client

func init() {
	config := AppConfig{}
	configor.New(&configor.Config{ENVPrefix: "SYNOLOGY"}).Load(&config)
	if len(config.Private.AccessId) == 0 {
		return
	}

	client = synology.New(&synology.Config{
		AccessId:  config.Public.AccessId,
		AccessKey: config.Public.AccessKey,
		Endpoint:  config.Public.Endpoint,
	})
	privateClient = synology.New(&synology.Config{
		AccessId:  config.Private.AccessId,
		AccessKey: config.Private.AccessKey,
		Endpoint:  config.Private.Endpoint,
	})
}

func TestAll(t *testing.T) {
	if client == nil {
		t.Skip(`skip because of no config:


			`)
	}
	clis := []*synology.Client{client, privateClient}
	for _, cli := range clis {
		tests.TestAll(cli, t)
	}
}
