// Package huawei 华为云OBS存储服务实现
// 提供华为云OBS的存储接口实现
package huawei

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	obs "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/smart-unicom/oss"
)

// 确保Client实现了StorageInterface接口
var _ oss.StorageInterface = (*Client)(nil)

// Client 华为云OBS存储客户端
// 封装华为云OBS的操作接口
type Client struct {
	// Config 客户端配置信息
	Config *Config
	// OBS 华为云OBS客户端实例
	OBS *obs.ObsClient
}

// Config 华为云OBS客户端配置
// 包含连接华为云OBS所需的所有配置参数
type Config struct {
	// SecretID 访问密钥ID
	SecretID string
	// SecretKey 访问密钥Secret
	SecretKey string
	// Endpoint 服务端点
	Endpoint string
	// Region 区域
	Region string
	// Bucket 存储桶名称
	Bucket string
	// SecurityToken 安全令牌（可选，用于临时访问凭证）
	SecurityToken string
}

// New 初始化华为云OBS存储客户端
// 参数:
//   - config: 华为云OBS配置信息
//
// 返回:
//   - *Client: 华为云OBS存储客户端实例
func New(config *Config) *Client {
	// 创建OBS客户端
	obsClient, err := obs.New(config.SecretID, config.SecretKey, config.Endpoint)
	if err != nil {
		panic(err)
	}

	return &Client{
		Config: config,
		OBS:    obsClient,
	}
}

// Get 获取指定路径的文件
// 参数:
//   - path: 文件路径
//
// 返回:
//   - *os.File: 文件对象
//   - error: 错误信息
func (client Client) Get(path string) (file *os.File, err error) {
	// 获取文件流
	readCloser, err := client.GetStream(path)
	if err != nil {
		return nil, err
	}

	// 创建临时文件并复制内容
	if file, err = os.CreateTemp("/tmp", "huaweicloud"); err == nil {
		defer readCloser.Close()
		// 将流内容复制到临时文件
		_, err = io.Copy(file, readCloser)
		// 重置文件指针到开始位置
		file.Seek(0, 0)
	}

	return file, err
}

// GetStream 获取指定路径文件的流
// 参数:
//   - path: 文件路径
//
// 返回:
//   - io.ReadCloser: 可读流
//   - error: 错误信息
func (client Client) GetStream(path string) (io.ReadCloser, error) {
	// 构建获取对象请求
	input := &obs.GetObjectInput{}
	input.Bucket = client.Config.Bucket
	input.Key = client.ToRelativePath(path)

	// 使用OBS客户端获取对象
	output, err := client.OBS.GetObject(input)
	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

// Put 上传文件到指定路径
// 参数:
//   - urlPath: 目标路径
//   - reader: 文件内容读取器
//
// 返回:
//   - *oss.Object: 上传后的对象信息
//   - error: 错误信息
func (client Client) Put(urlPath string, reader io.Reader) (*oss.Object, error) {
	// 如果是可寻址的读取器，重置到开始位置
	if seeker, ok := reader.(io.ReadSeeker); ok {
		seeker.Seek(0, 0)
	}

	// 构建上传对象请求
	input := &obs.PutObjectInput{}
	input.Bucket = client.Config.Bucket
	input.Key = client.ToRelativePath(urlPath)
	input.Body = reader

	// 使用OBS客户端上传对象
	_, err := client.OBS.PutObject(input)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &oss.Object{
		Path:             urlPath,
		Name:             filepath.Base(urlPath),
		LastModified:     &now,
		StorageInterface: client,
	}, nil
}

// Delete 删除指定路径的文件
// 参数:
//   - path: 文件路径
//
// 返回:
//   - error: 错误信息
func (client Client) Delete(path string) error {
	// 构建删除对象请求
	input := &obs.DeleteObjectInput{}
	input.Bucket = client.Config.Bucket
	input.Key = client.ToRelativePath(path)

	// 使用OBS客户端删除对象
	_, err := client.OBS.DeleteObject(input)
	return err
}

// List 列出指定路径下的所有对象
// 参数:
//   - path: 目录路径
//
// 返回:
//   - []*oss.Object: 对象列表
//   - error: 错误信息
func (client Client) List(path string) ([]*oss.Object, error) {
	var objects []*oss.Object

	// 构建列出对象请求
	input := &obs.ListObjectsInput{}
	input.Bucket = client.Config.Bucket
	input.Prefix = client.ToRelativePath(path)

	// 使用OBS客户端列出对象
	output, err := client.OBS.ListObjects(input)
	if err != nil {
		return nil, err
	}

	// 遍历对象列表并转换为统一格式
	for _, obj := range output.Contents {
		objects = append(objects, &oss.Object{
			Path:             "/" + obj.Key,
			Name:             filepath.Base(obj.Key),
			LastModified:     &obj.LastModified,
			Size:             obj.Size,
			StorageInterface: client,
		})
	}

	return objects, nil
}

// GetEndpoint 获取存储服务的端点地址
// 返回:
//   - string: 端点地址
func (client Client) GetEndpoint() string {
	// 返回华为云OBS的端点地址
	return client.Config.Endpoint
}

// ToRelativePath 将路径转换为相对路径
// 参数:
//   - urlPath: 原始路径
//
// 返回:
//   - string: 相对路径
func (client Client) ToRelativePath(urlPath string) string {
	// 移除路径前缀的斜杠
	return strings.TrimPrefix(urlPath, "/")
}

// GetURL 获取指定路径文件的访问URL
// 参数:
//   - path: 文件路径
//
// 返回:
//   - string: 访问URL
//   - error: 错误信息
func (client Client) GetURL(path string) (string, error) {
	// 构建生成预签名URL请求
	input := &obs.CreateSignedUrlInput{}
	input.Method = obs.HttpMethodGet
	input.Bucket = client.Config.Bucket
	input.Key = client.ToRelativePath(path)
	input.Expires = 3600 // 1小时有效期

	// 生成预签名URL
	output, err := client.OBS.CreateSignedUrl(input)
	if err != nil {
		return "", err
	}

	return output.SignedUrl, nil
}
