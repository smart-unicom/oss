# Azure Blob Storage

[Azure Blob Storage](https://azure.microsoft.com/zh-cn/services/storage/blobs/) 的存储后端实现

## 使用方法

```go
import "github.com/smart-unicom/oss/azureblob"

func main() {
  storage := azureblob.New(&azureblob.Config{
    AccountName:   "your_account_name",
    AccountKey:    "your_account_key",
    ContainerName: "your_container_name",
    Endpoint:      "https://your_account.blob.core.windows.net",
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

- `AccountName`: Azure存储账户名称
- `AccountKey`: Azure存储账户密钥
- `ContainerName`: Blob容器名称
- `Endpoint`: Azure Blob存储端点

## 环境变量配置

测试时可以通过以下环境变量配置：

```bash
export AZURE_ACCOUNT_NAME="your_account_name"
export AZURE_ACCOUNT_KEY="your_account_key"
export AZURE_CONTAINER_NAME="your_container_name"
export AZURE_ENDPOINT="https://your_account.blob.core.windows.net"
```

## 运行测试

```bash
go test ./azureblob
```