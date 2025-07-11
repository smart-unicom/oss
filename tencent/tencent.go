// Package tencent 腾讯云COS存储服务实现
// 提供腾讯云COS的存储接口实现
package tencent

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/smart-unicom/oss"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// 确保Client实现了StorageInterface接口
var _ oss.StorageInterface = (*Client)(nil)

// Config 腾讯云COS客户端配置
// 包含连接腾讯云COS所需的所有配置参数
type Config struct {
	// AppID 应用ID
	AppID string
	// SecretID 密钥ID
	SecretID string
	// SecretKey 密钥Key
	SecretKey string
	// Region 区域
	Region string
	// Bucket 存储桶名称
	Bucket string
	// ACL 访问权限控制列表
	ACL string
	// CORS 跨域资源共享
	CORS string
	// Endpoint 服务端点
	Endpoint string
}

// Client 腾讯云COS存储客户端
// 封装腾讯云COS的操作接口
type Client struct {
	// Config 客户端配置信息
	Config *Config
	// COS 腾讯云COS客户端实例
	COS *cos.Client
}

// New 初始化腾讯云COS存储客户端
// 参数:
//   - config: 腾讯云COS配置信息
//
// 返回:
//   - *Client: 腾讯云COS存储客户端实例
func New(config *Config) *Client {
	// 构建存储桶URL
	bucketURL := fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", config.Bucket, config.AppID, config.Region)
	u, _ := url.Parse(bucketURL)

	// 创建COS客户端
	cosClient := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretID,
			SecretKey: config.SecretKey,
		},
	})

	return &Client{
		Config: config,
		COS:    cosClient,
	}
}

// getUrl 获取腾讯云COS的访问URL
// 参数:
//   - path: 文件路径
//
// 返回:
//   - string: 访问URL
func (client Client) getUrl(path string) string {
	// 构建完整的COS访问URL
	return fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com/%s", client.Config.Bucket, client.Config.AppID, client.Config.Region, client.ToRelativePath(path))
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
	if file, err = ioutil.TempFile("/tmp", "tencent"); err == nil {
		defer readCloser.Close()
		// 将流内容复制到临时文件
		_, err = io.Copy(file, readCloser)
		// 重置文件指针到开始位置
		file.Seek(0, 0)
	}

	return file, err
}

// urlRegexp URL正则表达式，用于匹配HTTP/HTTPS URL
var urlRegexp = regexp.MustCompile(`(https?:)?//((\\w+).)+(\w+)/`)

// ToRelativePath 将路径转换为相对路径
// 参数:
//   - urlPath: 原始路径
//
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

// GetStream 获取指定路径文件的流
// 参数:
//   - path: 文件路径
//
// 返回:
//   - io.ReadCloser: 可读流
//   - error: 错误信息
func (client Client) GetStream(path string) (io.ReadCloser, error) {
	// 使用COS客户端获取对象
	resp, err := client.COS.Object.Get(context.Background(), client.ToRelativePath(path), nil)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// Put 上传文件到指定路径
// 参数:
//   - path: 目标路径
//   - body: 文件内容读取器
//
// 返回:
//   - *oss.Object: 上传后的对象信息
//   - error: 错误信息
func (client Client) Put(path string, body io.Reader) (*oss.Object, error) {
	// 如果是可寻址的读取器，重置到开始位置
	if seeker, ok := body.(io.ReadSeeker); ok {
		seeker.Seek(0, 0)
	}

	// 使用COS客户端上传对象
	_, err := client.COS.Object.Put(context.Background(), client.ToRelativePath(path), body, nil)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &oss.Object{
		Path:             path,
		Name:             filepath.Base(path),
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
	// 使用COS客户端删除对象
	_, err := client.COS.Object.Delete(context.Background(), client.ToRelativePath(path))
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

	// 使用COS客户端列出对象
	opt := &cos.BucketGetOptions{
		Prefix: client.ToRelativePath(path),
	}

	resp, _, err := client.COS.Bucket.Get(context.Background(), opt)
	if err != nil {
		return nil, err
	}

	// 遍历对象列表并转换为统一格式
	for _, obj := range resp.Contents {
		objects = append(objects, &oss.Object{
			Path: "/" + obj.Key,
			Name: filepath.Base(obj.Key),
			//LastModified:     &obj.LastModified,
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
	if client.Config.Endpoint != "" {
		return client.Config.Endpoint
	}
	// 返回腾讯云COS的标准端点格式
	return fmt.Sprintf("%s-%s.cos.%s.myqcloud.com", client.Config.Bucket, client.Config.AppID, client.Config.Region)
}

// GetURL 获取指定路径文件的访问URL
// 参数:
//   - path: 文件路径
//
// 返回:
//   - string: 访问URL
//   - error: 错误信息
func (client Client) GetURL(path string) (string, error) {
	// 返回文件的完整访问URL
	return client.getUrl(path), nil
}

// authorization 生成腾讯云COS的授权签名
// 参数:
//   - req: HTTP请求对象
//
// 返回:
//   - string: 授权签名字符串
func (client Client) authorization(req *http.Request) string {
	// 获取签名时间
	signTime := getSignTime()
	// 生成签名
	signature := getSignature(client.Config.SecretKey, req, signTime)
	// 构建授权字符串
	authStr := fmt.Sprintf("q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=%s&q-url-param-list=%s&q-signature=%s",
		client.Config.SecretID, signTime, signTime, getHeadKeys(req.Header), getParamsKeys(req.URL.RawQuery), signature)

	return authStr
}
