// Package aliyun 阿里云OSS存储服务实现
// 提供阿里云OSS的存储接口实现
package aliyun

import (
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	aliyun "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/smart-unicom/oss"
)

// Client 阿里云OSS存储客户端
// 封装阿里云OSS的操作接口
type Client struct {
	// Bucket OSS存储桶实例
	*aliyun.Bucket
	// Config 客户端配置信息
	Config *Config
}

// Config 阿里云OSS客户端配置
// 包含连接阿里云OSS所需的所有配置参数
type Config struct {
	// AccessId 访问密钥ID
	AccessId string
	// AccessKey 访问密钥Secret
	AccessKey string
	// Region 区域
	Region string
	// Bucket 存储桶名称
	Bucket string
	// Endpoint 服务端点
	Endpoint string
	// ACL 访问控制列表
	ACL aliyun.ACLType
	// ClientOptions 客户端选项
	ClientOptions []aliyun.ClientOption
	// UseCname 是否使用自定义域名
	UseCname bool
}

// New 初始化阿里云OSS存储客户端
// 参数:
//   - config: 阿里云OSS配置信息
// 返回:
//   - *Client: 阿里云OSS存储客户端实例
func New(config *Config) *Client {
	var (
		err    error
		client = &Client{Config: config}
	)

	// 设置默认端点
	if config.Endpoint == "" {
		config.Endpoint = "oss-cn-hangzhou.aliyuncs.com"
	}

	// 设置默认访问控制
	if config.ACL == "" {
		config.ACL = aliyun.ACLPublicRead
	}

	// 配置自定义域名
	if config.UseCname {
		config.ClientOptions = append(config.ClientOptions, aliyun.UseCname(config.UseCname))
	}

	// 创建阿里云OSS客户端
	Aliyun, err := aliyun.New(config.Endpoint, config.AccessId, config.AccessKey, config.ClientOptions...)

	if err == nil {
		// 获取存储桶实例
		client.Bucket, err = Aliyun.Bucket(config.Bucket)
	}

	if err != nil {
		panic(err)
	}

	return client
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

	// 创建临时文件并复制内容
	if file, err = ioutil.TempFile("/tmp", "ali"); err == nil {
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
// 返回:
//   - io.ReadCloser: 可读流
//   - error: 错误信息
func (client Client) GetStream(path string) (io.ReadCloser, error) {
	// 从OSS获取对象流
	return client.Bucket.GetObject(client.ToRelativePath(path))
}

// Put 上传文件到指定路径
// 参数:
//   - urlPath: 目标路径
//   - reader: 文件内容读取器
// 返回:
//   - *oss.Object: 上传后的对象信息
//   - error: 错误信息
func (client Client) Put(urlPath string, reader io.Reader) (*oss.Object, error) {
	// 如果是可寻址的读取器，重置到开始位置
	if seeker, ok := reader.(io.ReadSeeker); ok {
		seeker.Seek(0, 0)
	}

	// 上传对象到阿里云OSS
	err := client.Bucket.PutObject(client.ToRelativePath(urlPath), reader, aliyun.ACL(client.Config.ACL))
	now := time.Now()

	return &oss.Object{
		Path:             urlPath,
		Name:             filepath.Base(urlPath),
		LastModified:     &now,
		StorageInterface: client,
	}, err
}

// Delete 删除指定路径的文件
// 参数:
//   - path: 文件路径
// 返回:
//   - error: 错误信息
func (client Client) Delete(path string) error {
	return client.Bucket.DeleteObject(client.ToRelativePath(path))
}

// List 列出指定路径下的所有对象
// 参数:
//   - path: 目录路径
// 返回:
//   - []*oss.Object: 对象列表
//   - error: 错误信息
func (client Client) List(path string) ([]*oss.Object, error) {
	var objects []*oss.Object

	// 列出指定前缀的所有对象
	results, err := client.Bucket.ListObjects(aliyun.Prefix(path))

	if err == nil {
		// 遍历结果并转换为统一的对象格式
		for _, obj := range results.Objects {
			objects = append(objects, &oss.Object{
				Path:             "/" + client.ToRelativePath(obj.Key),
				Name:             filepath.Base(obj.Key),
				LastModified:     &obj.LastModified,
				Size:             obj.Size,
				StorageInterface: client,
			})
		}
	}

	return objects, err
}

// GetEndpoint 获取存储服务的端点地址
// 返回:
//   - string: 端点地址
func (client Client) GetEndpoint() string {
	if client.Config.Endpoint != "" {
		// 如果是阿里云标准域名，添加存储桶前缀
		if strings.HasSuffix(client.Config.Endpoint, "aliyuncs.com") {
			return client.Config.Bucket + "." + client.Config.Endpoint
		}
		return client.Config.Endpoint
	}

	// 从客户端配置中获取端点
	endpoint := client.Bucket.Client.Config.Endpoint
	// 移除协议前缀
	for _, prefix := range []string{"https://", "http://"} {
		endpoint = strings.TrimPrefix(endpoint, prefix)
	}

	return client.Config.Bucket + "." + endpoint
}

// urlRegexp URL正则表达式，用于匹配HTTP/HTTPS URL
var urlRegexp = regexp.MustCompile(`(https?:)?//((\w+).)+(\w+)/`)

// ToRelativePath 将路径转换为相对路径
// 参数:
//   - urlPath: 原始路径
// 返回:
//   - string: 相对路径
func (client Client) ToRelativePath(urlPath string) string {
	// 如果是完整的URL，解析并提取路径部分
	if urlRegexp.MatchString(urlPath) {
		if u, err := url.Parse(urlPath); err == nil {
			return strings.TrimPrefix(u.Path, "/")
		}
	}

	// 移除路径前缀的斜杠
	return strings.TrimPrefix(urlPath, "/")
}

// GetURL 获取指定路径文件的访问URL
// 参数:
//   - path: 文件路径
// 返回:
//   - string: 访问URL
//   - error: 错误信息
func (client Client) GetURL(path string) (url string, err error) {
	// 如果是私有访问，生成签名URL（1小时有效期）
	if client.Config.ACL == aliyun.ACLPrivate {
		return client.Bucket.SignURL(client.ToRelativePath(path), aliyun.HTTPGet, 60*60)
	}
	// 公共访问直接返回路径
	return path, nil
}
