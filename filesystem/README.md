# 本地文件系统存储

本地文件系统的存储后端实现，用于开发测试或本地文件管理

## 使用方法

```go
import "github.com/smart-unicom/oss/filesystem"

func main() {
  storage := filesystem.New(&filesystem.Config{
    RootPath: "/path/to/storage/root",
    BaseURL:  "http://localhost:8080/files", // 可选，用于生成访问URL
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

- `RootPath`: 本地存储根目录路径
- `BaseURL`: 基础URL，用于生成文件访问链接（可选）

## 特性

- **本地存储**: 文件直接存储在本地文件系统中
- **开发友好**: 适合开发环境和测试使用
- **零依赖**: 不需要外部服务，直接使用操作系统文件API
- **路径管理**: 自动创建目录结构

## 环境变量配置

测试时可以通过以下环境变量配置：

```bash
export FILESYSTEM_ROOT_PATH="/tmp/oss_test"
export FILESYSTEM_BASE_URL="http://localhost:8080/files"
```

## 运行测试

```bash
go test ./filesystem
```

## 注意事项

- 确保指定的根目录路径存在且有读写权限
- 在生产环境中使用时，注意文件权限和安全性
- BaseURL配置影响GetURL方法返回的链接格式