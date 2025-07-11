// Package googlecloud Google Cloud存储服务实现
// 提供Google Cloud Storage的存储接口实现
package googlecloud

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/smart-unicom/oss"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client Google Cloud存储客户端
// 封装Google Cloud Storage的操作接口
type Client struct {
	// Config 客户端配置信息
	Config *Config
	// BucketHandle 存储桶句柄
	BucketHandle *storage.BucketHandle
}

// Config Google Cloud客户端配置
// 包含连接Google Cloud Storage所需的所有配置参数
type Config struct {
	// ServiceAccountJson 服务账户JSON密钥
	ServiceAccountJson string
	// Bucket 存储桶名称
	Bucket string
	// Endpoint 服务端点
	Endpoint string
}

// New 初始化Google Cloud存储客户端
// 参数:
//   - config: Google Cloud配置信息
// 返回:
//   - *Client: Google Cloud存储客户端实例
//   - error: 错误信息
func New(config *Config) (*Client, error) {
	// 创建上下文
	ctx := context.Background()
	// 从JSON创建凭据
	credentials, err := google.CredentialsFromJSON(ctx, []byte(config.ServiceAccountJson), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}

	// 创建存储客户端
	storageClient, err := storage.NewClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, err
	}

	// 创建客户端实例
	client := &Client{
		Config:       config,
		BucketHandle: storageClient.Bucket(config.Bucket),
	}
	return client, nil
}

// Get 获取指定路径的文件
// 参数:
//   - path: 文件路径
// 返回:
//   - *os.File: 文件对象
//   - error: 错误信息
func (client Client) Get(path string) (file *os.File, err error) {
	// 获取文件流
	readCloser, err := client.GetStream(path)
	if err != nil {
		return nil, err
	}
	defer readCloser.Close()

	// 创建临时文件
	file, err = ioutil.TempFile("/tmp", "googlecloud")
	if err != nil {
		return nil, err
	}

	// 将流内容复制到临时文件
	_, err = io.Copy(file, readCloser)
	if err != nil {
		return nil, err
	}

	// 重置文件指针到开始位置
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// GetStream 获取指定路径文件的流
// 参数:
//   - path: 文件路径
// 返回:
//   - io.ReadCloser: 可读流
//   - error: 错误信息
func (client Client) GetStream(path string) (io.ReadCloser, error) {
	// 创建上下文
	ctx := context.Background()
	// 检查对象是否存在
	_, err := client.BucketHandle.Object(path).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	// 创建对象读取器
	return client.BucketHandle.Object(path).NewReader(ctx)
}

// Put 上传文件到指定路径
// 参数:
//   - urlPath: 目标路径
//   - reader: 文件内容读取器
// 返回:
//   - *oss.Object: 上传后的对象信息
//   - error: 错误信息
func (client Client) Put(urlPath string, reader io.Reader) (*oss.Object, error) {
	// 创建上下文
	ctx := context.Background()

	// 创建对象写入器
	wc := client.BucketHandle.Object(urlPath).NewWriter(ctx)

	// 将内容复制到写入器
	_, err := io.Copy(wc, reader)
	if err != nil {
		return nil, err
	}

	// 关闭写入器以完成上传
	err = wc.Close()
	if err != nil {
		return nil, err
	}

	// 获取对象属性
	attrs, err := client.BucketHandle.Object(urlPath).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	// 创建返回对象
	res := &oss.Object{
		Path:             urlPath,
		Name:             filepath.Base(urlPath),
		LastModified:     &attrs.Updated,
		StorageInterface: client,
	}
	return res, nil
}

// Delete 删除指定路径的文件
// 参数:
//   - path: 文件路径
// 返回:
//   - error: 错误信息
func (client Client) Delete(path string) error {
	// 创建上下文并删除对象
	ctx := context.Background()
	return client.BucketHandle.Object(path).Delete(ctx)
}

// List 列出指定路径下的所有对象
// 参数:
//   - path: 路径前缀
// 返回:
//   - []*oss.Object: 对象列表
//   - error: 错误信息
func (client Client) List(path string) ([]*oss.Object, error) {
	var objects []*oss.Object
	// 创建上下文
	ctx := context.Background()

	// 创建对象迭代器
	iter := client.BucketHandle.Objects(ctx, &storage.Query{Prefix: path})
	for {
		// 获取下一个对象属性
		objAttrs, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// 添加到对象列表
		objects = append(objects, &oss.Object{
			Path:             "/" + objAttrs.Name,
			Name:             filepath.Base(objAttrs.Name),
			LastModified:     &objAttrs.Updated,
			Size:             objAttrs.Size,
			StorageInterface: client,
		})
	}

	return objects, nil
}

// GetURL 获取指定路径文件的访问URL
// 参数:
//   - path: 文件路径
// 返回:
//   - string: 访问URL
//   - error: 错误信息
func (client Client) GetURL(path string) (url string, err error) {
	return path, nil
}

// GetEndpoint 获取存储服务的端点地址
// 返回:
//   - string: 端点地址
func (client Client) GetEndpoint() string {
	// 如果配置了自定义端点，使用自定义端点
	if client.Config.Endpoint != "" {
		return client.Config.Endpoint
	}
	// 返回Google Cloud Storage的默认端点
	return "https://storage.googleapis.com"
}

// ToRelativePath 将路径转换为相对路径
// 参数:
//   - urlPath: 原始路径
// 返回:
//   - string: 相对路径
func (client Client) ToRelativePath(urlPath string) string {
	// 如果路径包含端点前缀，移除它
	if strings.HasPrefix(urlPath, client.GetEndpoint()) {
		relativePath := strings.TrimPrefix(urlPath, client.GetEndpoint())
		return strings.TrimPrefix(relativePath, "/")
	}
	return urlPath
}
