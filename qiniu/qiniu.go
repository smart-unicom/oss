// Package qiniu 七牛云对象存储服务实现
// 提供七牛云Kodo的存储接口实现
package qiniu

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/smart-unicom/oss"
)

// Client 七牛云存储客户端
// 封装七牛云Kodo的操作接口
type Client struct {
	// Config 客户端配置信息
	Config *Config
	// mac 七牛云认证管理器
	mac *qbox.Mac
	// storageCfg 存储配置
	storageCfg storage.Config
	// bucketManager 存储桶管理器
	bucketManager *storage.BucketManager
	// putPolicy 上传策略
	putPolicy *storage.PutPolicy
}

// Config 七牛云客户端配置
// 包含连接七牛云Kodo所需的所有配置参数
type Config struct {
	// AccessId 访问密钥ID
	AccessId string
	// AccessKey 访问密钥
	AccessKey string
	// Region 地域
	Region string
	// Bucket 存储桶名称
	Bucket string
	// Endpoint 服务端点
	Endpoint string
	// UseHTTPS 是否使用HTTPS
	UseHTTPS bool
	// UseCdnDomains 是否使用CDN域名
	UseCdnDomains bool
	// PrivateURL 是否为私有URL
	PrivateURL bool
}

// zonedata 七牛云存储区域映射表
// 将字符串区域名称映射到七牛云的Zone对象
var zonedata = map[string]*storage.Zone{
	"huadong": &storage.ZoneHuadong, // 华东区域
	"huabei":  &storage.ZoneHuabei,  // 华北区域
	"huanan":  &storage.ZoneHuanan,  // 华南区域
	"beimei":  &storage.ZoneBeimei,  // 北美区域
}

// New 初始化七牛云存储客户端
// 参数:
//   - config: 七牛云配置信息
//
// 返回:
//   - *Client: 七牛云存储客户端实例
//   - error: 错误信息
func New(config *Config) (*Client, error) {
	// 创建客户端实例
	client := &Client{Config: config, storageCfg: storage.Config{}}

	// 初始化认证管理器
	client.mac = qbox.NewMac(config.AccessId, config.AccessKey)

	// 设置存储区域
	if z, ok := zonedata[strings.ToLower(config.Region)]; ok {
		client.storageCfg.Zone = z
	} else {
		return nil, fmt.Errorf("Zone %s is invalid, only support huadong, huabei, huanan, beimei.", config.Region)
	}

	// 验证端点配置
	if len(config.Endpoint) == 0 {
		return nil, fmt.Errorf("endpoint must be provided.")
	}

	// 验证端点格式
	if !strings.HasPrefix(config.Endpoint, "http://") && !strings.HasPrefix(config.Endpoint, "https://") {
		return nil, fmt.Errorf("endpoint must start with http:// or https://")
	}

	// 配置存储选项
	client.storageCfg.UseHTTPS = config.UseHTTPS
	client.storageCfg.UseCdnDomains = config.UseCdnDomains

	// 初始化存储桶管理器
	client.bucketManager = storage.NewBucketManager(client.mac, &client.storageCfg)

	return client, nil
}

// SetPutPolicy 设置上传策略
// 参数:
//   - putPolicy: 七牛云上传策略
func (client Client) SetPutPolicy(putPolicy *storage.PutPolicy) {
	client.putPolicy = putPolicy
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

	// 创建临时文件并复制内容
	if file, err = ioutil.TempFile(os.TempDir(), "qiniu"); err == nil {
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
	// 获取文件的访问URL
	purl, err := client.GetURL(path)
	if err != nil {
		return nil, err
	}

	// 发送HTTP GET请求获取文件
	var res *http.Response
	res, err = http.Get(purl)
	if err == nil && res.StatusCode != http.StatusOK {
		err = fmt.Errorf("file %s not found", path)
	}

	return res.Body, err
}

// Put 上传文件到指定路径
// 参数:
//   - urlPath: 文件路径
//   - reader: 文件内容读取器
//
// 返回:
//   - *oss.Object: 上传成功后的对象信息
//   - error: 错误信息
func (client Client) Put(urlPath string, reader io.Reader) (r *oss.Object, err error) {
	// 如果reader支持Seek，重置到开始位置
	if seeker, ok := reader.(io.ReadSeeker); ok {
		seeker.Seek(0, 0)
	}

	// 处理存储键
	urlPath = storageKey(urlPath)
	var buffer []byte
	buffer, err = ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	// 检测文件类型
	fileType := mime.TypeByExtension(path.Ext(urlPath))
	if fileType == "" {
		fileType = http.DetectContentType(buffer)
	}

	// 设置上传策略
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", client.Config.Bucket, urlPath),
	}

	// 如果客户端有自定义上传策略，使用自定义策略
	if client.putPolicy != nil {
		putPolicy = *client.putPolicy
	}

	// 生成上传凭证
	upToken := putPolicy.UploadToken(client.mac)

	// 创建表单上传器
	formUploader := storage.NewFormUploader(&client.storageCfg)
	ret := storage.PutRet{}
	dataLen := int64(len(buffer))

	// 设置上传参数
	putExtra := storage.PutExtra{
		Params: map[string]string{},
	}
	// 执行文件上传
	err = formUploader.Put(context.Background(), &ret, upToken, urlPath, bytes.NewReader(buffer), dataLen, &putExtra)
	if err != nil {
		return
	}

	// 创建返回对象
	now := time.Now()
	return &oss.Object{
		Path:             ret.Key,
		Name:             filepath.Base(urlPath),
		LastModified:     &now,
		StorageInterface: client,
	}, err
}

// Delete 删除指定路径的文件
// 参数:
//   - path: 文件路径
//
// 返回:
//   - error: 错误信息
func (client Client) Delete(path string) error {
	return client.bucketManager.Delete(client.Config.Bucket, storageKey(path))
}

// List 列出指定路径下的所有对象
// 参数:
//   - path: 路径前缀
//
// 返回:
//   - []*oss.Object: 对象列表
//   - error: 错误信息
func (client Client) List(path string) (objects []*oss.Object, err error) {
	// 处理路径前缀
	var prefix = storageKey(path)
	var listItems []storage.ListItem
	// 获取文件列表
	listItems, _, _, _, err = client.bucketManager.ListFiles(
		client.Config.Bucket,
		prefix,
		"",
		"",
		100,
	)

	if err != nil {
		return
	}

	// 转换为oss.Object格式
	for _, content := range listItems {
		t := time.Unix(content.PutTime, 0)
		objects = append(objects, &oss.Object{
			Path:             "/" + storageKey(content.Key),
			Name:             filepath.Base(content.Key),
			LastModified:     &t,
			StorageInterface: client,
		})
	}

	return
}

// GetEndpoint 获取存储端点
// 返回:
//   - string: 存储端点URL
func (client Client) GetEndpoint() string {
	return client.Config.Endpoint
}

var urlRegexp = regexp.MustCompile(`(https?:)?//((\w+).)+(\w+)/`)

// storageKey 处理存储键，去除URL前缀并标准化路径
// 参数:
//   - urlPath: 原始URL路径
//
// 返回:
//   - string: 处理后的存储键
func storageKey(urlPath string) string {
	if urlRegexp.MatchString(urlPath) {
		if u, err := url.Parse(urlPath); err == nil {
			urlPath = u.Path
		}
	}
	return strings.TrimPrefix(urlPath, "/")
}

// GetURL 获取文件的公共访问URL
// 参数:
//   - path: 文件路径
//
// 返回:
//   - string: 公共访问URL
//   - error: 错误信息
func (client Client) GetURL(path string) (url string, err error) {
	if len(path) == 0 {
		return
	}
	key := storageKey(path)

	// 如果配置为私有URL，生成带签名的私有访问URL
	if client.Config.PrivateURL {
		deadline := time.Now().Add(time.Second * 3600).Unix()
		url = storage.MakePrivateURL(client.mac, client.Config.Endpoint, key, deadline)
		return
	}

	// 生成公共访问URL
	url = storage.MakePublicURL(client.GetEndpoint(), key)

	return
}
