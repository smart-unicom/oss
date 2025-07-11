# Google Cloud Storage

[Google Cloud Storage](https://cloud.google.com/storage/) 的存储后端实现

## 使用方法

```go
import "github.com/smart-unicom/oss/googlecloud"

func main() {
  storage := googlecloud.New(&googlecloud.Config{
    Bucket:           "your_bucket_name",
    ProjectID:        "your_project_id",
    CredentialsFile:  "/path/to/service-account.json", // 可选
    Endpoint:         "https://storage.googleapis.com", // 可选
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

- `Bucket`: Google Cloud Storage存储桶名称
- `ProjectID`: Google Cloud项目ID
- `CredentialsFile`: 服务账户JSON密钥文件路径（可选）
- `Endpoint`: 自定义端点（可选）

## 认证方式

### 1. 服务账户密钥文件
```go
storage := googlecloud.New(&googlecloud.Config{
  Bucket:          "your_bucket_name",
  ProjectID:       "your_project_id",
  CredentialsFile: "/path/to/service-account.json",
})
```

### 2. 环境变量认证
```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
```

### 3. 默认凭据（在GCP环境中）
在Google Cloud环境中运行时，会自动使用默认凭据。

## 环境变量配置

测试时可以通过以下环境变量配置：

```bash
export GOOGLE_CLOUD_PROJECT="your_project_id"
export GOOGLE_CLOUD_BUCKET="your_bucket_name"
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
export GOOGLE_CLOUD_ENDPOINT="https://storage.googleapis.com"
```

## 运行测试

```bash
go test ./googlecloud
```

## 注意事项

- 确保服务账户具有相应的Cloud Storage权限
- 存储桶名称必须全局唯一
- 建议在生产环境中使用IAM角色而非密钥文件

