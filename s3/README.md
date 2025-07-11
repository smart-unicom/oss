# AWS S3

[AWS S3](https://aws.amazon.com/cn/s3/) 的存储后端实现

## 使用方法

```go
import "github.com/smart-unicom/oss/s3"

func main() {
  storage := s3.New(&s3.Config{
    AccessID:  "your_access_key_id",
    AccessKey: "your_secret_access_key",
    Region:    "us-west-2",
    Bucket:    "your_bucket_name",
    Endpoint:  "s3.amazonaws.com", // 可选，自定义端点
    ACL:       "public-read",      // 可选，访问控制列表
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

- `AccessID`: AWS访问密钥ID
- `AccessKey`: AWS访问密钥Secret
- `Region`: AWS区域，如 `us-west-2`、`ap-northeast-1`
- `Bucket`: S3存储桶名称
- `Endpoint`: 自定义端点（可选）
- `ACL`: 访问控制列表（可选）

## 常用区域

- `us-east-1`: 美国东部（弗吉尼亚北部）
- `us-west-2`: 美国西部（俄勒冈）
- `ap-northeast-1`: 亚太地区（东京）
- `ap-southeast-1`: 亚太地区（新加坡）
- `eu-west-1`: 欧洲（爱尔兰）

## 环境变量配置

测试时可以通过以下环境变量配置：

```bash
export AWS_ACCESS_KEY_ID="your_access_key_id"
export AWS_SECRET_ACCESS_KEY="your_secret_access_key"
export AWS_REGION="us-west-2"
export AWS_BUCKET="your_bucket_name"
export AWS_ENDPOINT="s3.amazonaws.com"
```

## 运行测试

```bash
go test ./s3
```


