# 阿里云OSS

[阿里云对象存储服务(OSS)](https://www.aliyun.com/product/oss) 的存储后端实现

## 使用方法

```go
import "github.com/smart-unicom/oss/aliyun"

func main() {
  storage := aliyun.New(&aliyun.Config{
    AccessKeyID:     "your_access_key_id",
    AccessKeySecret: "your_access_key_secret",
    Bucket:          "your_bucket_name",
    Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
    Region:          "cn-hangzhou", // 可选
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

- `AccessKeyID`: 阿里云访问密钥ID
- `AccessKeySecret`: 阿里云访问密钥Secret
- `Bucket`: OSS存储桶名称
- `Endpoint`: OSS服务端点
- `Region`: 区域代码（可选）

## 常用地域端点

- `oss-cn-hangzhou.aliyuncs.com`: 华东1（杭州）
- `oss-cn-shanghai.aliyuncs.com`: 华东2（上海）
- `oss-cn-beijing.aliyuncs.com`: 华北2（北京）
- `oss-cn-shenzhen.aliyuncs.com`: 华南1（深圳）
- `oss-cn-hongkong.aliyuncs.com`: 香港

## 环境变量配置

测试时可以通过以下环境变量配置：

```bash
export ALIYUN_ACCESS_KEY_ID="your_access_key_id"
export ALIYUN_ACCESS_KEY_SECRET="your_access_key_secret"
export ALIYUN_BUCKET="your_bucket_name"
export ALIYUN_ENDPOINT="oss-cn-hangzhou.aliyuncs.com"
export ALIYUN_REGION="cn-hangzhou"
```

## 运行测试

```bash
go test ./aliyun
```
