# 腾讯云COS

[腾讯云对象存储(COS)](https://cloud.tencent.com/product/cos) 的存储后端实现

## 使用方法

```go
import "github.com/smart-unicom/oss/tencent"

func main() {
  storage := tencent.New(&tencent.Config{
    SecretID:  "your_secret_id",
    SecretKey: "your_secret_key",
    Region:    "ap-beijing",
    Bucket:    "your_bucket_name",
    AppID:     "your_app_id",
    BaseURL:   "https://your_bucket.cos.ap-beijing.myqcloud.com", // 可选
  })

  // 保存文件到存储
  storage.Put("/sample.txt", reader)

  // 根据路径获取文件
  storage.Get("/sample.txt")

  // 获取文件流
  storage.GetStream("/sample.txt")

  // 删除文件
  storage.Delete("/sample.txt")

  // 列出指定路径下的所有对象
  storage.List("/")

  // 获取公共访问URL
  storage.GetURL("/sample.txt")
}
```

## 配置说明

- `SecretID`: 腾讯云访问密钥ID
- `SecretKey`: 腾讯云访问密钥Secret
- `Region`: 存储桶所在地域，如 `ap-beijing`、`ap-shanghai`
- `Bucket`: 存储桶名称
- `AppID`: 腾讯云应用ID
- `BaseURL`: 自定义域名或CDN域名（可选）

## 地域列表

常用地域代码：
- `ap-beijing`: 北京
- `ap-shanghai`: 上海
- `ap-guangzhou`: 广州
- `ap-chengdu`: 成都
- `ap-hongkong`: 香港
- `ap-singapore`: 新加坡

## 环境变量配置

测试时可以通过以下环境变量配置：

```bash
export TENCENT_SECRET_ID="your_secret_id"
export TENCENT_SECRET_KEY="your_secret_key"
export TENCENT_REGION="ap-beijing"
export TENCENT_BUCKET="your_bucket_name"
export TENCENT_APP_ID="your_app_id"
export TENCENT_BASE_URL="https://your_bucket.cos.ap-beijing.myqcloud.com"
```

## 运行测试

```bash
go test ./tencent
```

## 注意事项

- 存储桶名称格式：`bucket-appid`
- 确保SecretID和SecretKey具有相应的COS操作权限
- 不同地域的访问端点不同，请选择合适的地域