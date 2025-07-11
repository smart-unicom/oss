package huawei_test

import (
	"os"
	"testing"

	"github.com/smart-unicom/oss/huawei"
	"github.com/smart-unicom/oss/tests"
)

func TestHuawei(t *testing.T) {
	// 从环境变量获取华为云OBS配置
	config := &huawei.Config{
		SecretID:      os.Getenv("HUAWEI_SECRET_ID"),
		SecretKey:     os.Getenv("HUAWEI_SECRET_KEY"),
		Endpoint:      os.Getenv("HUAWEI_ENDPOINT"),
		Region:        os.Getenv("HUAWEI_REGION"),
		Bucket:        os.Getenv("HUAWEI_BUCKET"),
		SecurityToken: os.Getenv("HUAWEI_SECURITY_TOKEN"),
	}

	// 检查必要的配置是否存在
	if config.SecretID == "" || config.SecretKey == "" || config.Endpoint == "" || config.Bucket == "" {
		t.Skip("华为云OBS配置不完整，跳过测试")
		return
	}

	// 创建华为云OBS客户端
	client := huawei.New(config)

	// 运行通用测试
	tests.TestAll(client, t)
}
