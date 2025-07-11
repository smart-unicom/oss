# 七牛云对象存储

[七牛云对象存储](https://www.qiniu.com/products/kodo) 的存储后端实现

## 使用方法

```go
import "github.com/smart-unicom/oss/qiniu"

func main() {
  storage := qiniu.New(&qiniu.Config{
    AccessKey:    "your_access_key",
    SecretKey:    "your_secret_key",
    Bucket:       "your_bucket_name",
    Domain:       "http://your-domain.clouddn.com",
    Region:       "z0", // 华东区域
    UseHTTPS:     false, // 可选
    UseCdnDomains: false, // 可选
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

- `AccessKey`: 七牛云访问密钥
- `SecretKey`: 七牛云私钥
- `Bucket`: 存储空间名称
- `Domain`: 存储空间绑定的域名
- `Region`: 存储区域代码
- `UseHTTPS`: 是否使用HTTPS（可选）
- `UseCdnDomains`: 是否使用CDN加速域名（可选）

## 存储区域

- `z0`: 华东-浙江
- `z1`: 华北-河北
- `z2`: 华南-广东
- `na0`: 北美-洛杉矶
- `as0`: 亚太-新加坡
- `cn-east-2`: 华东-浙江2

## 环境变量配置

测试时可以通过以下环境变量配置：

```bash
export QINIU_ACCESS_KEY="your_access_key"
export QINIU_SECRET_KEY="your_secret_key"
export QINIU_BUCKET="your_bucket_name"
export QINIU_DOMAIN="http://your-domain.clouddn.com"
export QINIU_REGION="z0"
export QINIU_USE_HTTPS="false"
export QINIU_USE_CDN_DOMAINS="false"
```

## 运行测试

```bash
go test ./qiniu
```

## 注意事项

- 域名需要在七牛云控制台绑定到存储空间
- 建议在生产环境中使用HTTPS
- 可以根据业务需求选择合适的存储区域

