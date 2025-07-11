# OSS 对象存储组件库

Golang 统一的对象存储接口，支持多种云存储服务提供商,包括支持阿里云、腾讯云、华为云、七牛云对象存储、AWS S3、Google Cloud、Azure、本地文件系统

## 支持的存储服务

- [阿里云 OSS](./aliyun/README.md)
- [腾讯云 COS](./tencent/README.md)
- [华为云 OBS](./huaweicloud/README.md)
- [七牛云对象存储](./qiniu/README.md)
- [AWS S3](./s3/README.md)
- [Google Cloud Storage](./googlecloud/README.md)
- [Azure Blob Storage](./azureblob/README.md)
- [本地文件系统](./filesystem/README.md)

## 统一接口

所有存储后端都实现了相同的接口：

```go
type StorageInterface interface {
    Get(path string) (*os.File, error)
    GetStream(path string) (io.ReadCloser, error)
    Put(path string, reader io.Reader) (*Object, error)
    Delete(path string) error
    List(path string) ([]*Object, error)
    GetURL(path string) (string, error)
}
```

## 快速开始

### 华为云 OBS 示例

```go
import "github.com/smart-unicom/oss/huaweicloud"

func main() {
  storage := huaweicloud.New(&huaweicloud.Config{
    AccessKeyID:     "your_access_key_id",
    SecretAccessKey: "your_secret_access_key",
    Endpoint:        "obs.cn-north-4.myhuaweicloud.com",
    Region:          "cn-north-4",
    Bucket:          "your_bucket_name",
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

### 阿里云 OSS 示例

```go
import "github.com/smart-unicom/oss/aliyun"

storage := aliyun.New(&aliyun.Config{
  AccessKeyID:     "your_access_key_id",
  AccessKeySecret: "your_access_key_secret",
  Bucket:          "your_bucket_name",
  Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
})
```

### 腾讯云 COS 示例

```go
import "github.com/smart-unicom/oss/tencent"

storage := tencent.New(&tencent.Config{
  SecretID:  "your_secret_id",
  SecretKey: "your_secret_key",
  Bucket:    "your_bucket_name",
  Region:    "ap-beijing",
})
```

## 特性

- **统一接口**: 所有存储后端使用相同的API
- **多云支持**: 支持主流云存储服务
- **流式处理**: 支持大文件的流式上传和下载
- **URL生成**: 支持生成临时访问URL
- **路径管理**: 统一的路径处理机制
- **错误处理**: 完善的错误处理和日志记录
- **测试覆盖**: 每个后端都有完整的测试用例

## 安装

```bash
go get github.com/smart-unicom/oss
```

## 测试

运行所有测试：

```bash
go test ./...
```

运行特定存储后端的测试：

```bash
go test ./aliyun
go test ./tencent
go test ./huaweicloud
# ... 其他后端
```

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 许可证

MIT License
