// Package s3 提供AWS S3存储的实现
// 支持AWS S3存储服务的文件上传、下载、删除等操作
package s3

import (
	"bytes"
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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/smart-unicom/oss"
)

// Client AWS S3存储客户端
// 封装了AWS S3存储的操作接口
type Client struct {
	*s3.S3        // AWS S3服务客户端
	Config *Config // 配置信息
}

// Config AWS S3存储配置
// 包含连接AWS S3存储所需的所有配置信息
type Config struct {
	AccessId         string            // 访问密钥ID
	AccessKey        string            // 访问密钥
	Region           string            // AWS区域
	Bucket           string            // 存储桶名称
	SessionToken     string            // 会话令牌
	ACL              string            // 访问控制列表
	Endpoint         string            // 端点URL
	S3Endpoint       string            // S3端点URL
	S3ForcePathStyle bool              // 是否强制使用路径样式
	CacheControl     string            // 缓存控制

	Session *session.Session          // AWS会话

	RoleARN string                    // IAM角色ARN
}

// ec2RoleAwsCreds 获取EC2角色的AWS凭据
// 参数:
//   - config: S3配置信息
// 返回:
//   - *credentials.Credentials: AWS凭据对象
func ec2RoleAwsCreds(config *Config) *credentials.Credentials {
	// 创建EC2元数据客户端
	ec2m := ec2metadata.New(session.New(), &aws.Config{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Endpoint:   aws.String("http://169.254.169.254/latest"),
	})

	// 返回EC2角色凭据提供者
	return credentials.NewCredentials(&ec2rolecreds.EC2RoleProvider{
		Client: ec2m,
	})
}

// EC2RoleAwsConfig 创建使用EC2角色的AWS配置
// 参数:
//   - config: S3配置信息
// 返回:
//   - *aws.Config: AWS配置对象
func EC2RoleAwsConfig(config *Config) *aws.Config {
	return &aws.Config{
		Region:      aws.String(config.Region),
		Credentials: ec2RoleAwsCreds(config),
	}
}

// New 初始化S3存储客户端
// 参数:
//   - config: S3配置信息
// 返回:
//   - *Client: S3存储客户端实例
func New(config *Config) *Client {
	// 如果未设置ACL，使用默认的公共读取权限
	if config.ACL == "" {
		config.ACL = s3.BucketCannedACLPublicRead
	}

	// 创建客户端实例
	client := &Client{Config: config}

	// 如果配置了IAM角色ARN，使用STS凭据
	if config.RoleARN != "" {
		sess := session.Must(session.NewSession())
		creds := stscreds.NewCredentials(sess, config.RoleARN)

		s3Config := &aws.Config{
			Region:           &config.Region,
			Endpoint:         &config.S3Endpoint,
			S3ForcePathStyle: &config.S3ForcePathStyle,
			Credentials:      creds,
		}

		client.S3 = s3.New(sess, s3Config)
		return client
	}

	// 创建基础S3配置
	s3Config := &aws.Config{
		Region:           &config.Region,
		Endpoint:         &config.S3Endpoint,
		S3ForcePathStyle: &config.S3ForcePathStyle,
	}

	// 根据不同的认证方式初始化S3客户端
	if config.Session != nil {
		// 使用提供的会话
		client.S3 = s3.New(config.Session, s3Config)
	} else if config.AccessId == "" && config.AccessKey == "" {
		// 使用AWS默认凭据
		sess := session.Must(session.NewSession())
		client.S3 = s3.New(sess, s3Config)
	} else {
		// 使用静态凭据
		creds := credentials.NewStaticCredentials(config.AccessId, config.AccessKey, config.SessionToken)
		if _, err := creds.Get(); err == nil {
			s3Config.Credentials = creds
			client.S3 = s3.New(session.New(), s3Config)
		}
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

	// 根据文件扩展名生成临时文件模式
	ext := filepath.Ext(path)
	pattern := fmt.Sprintf("s3*%s", ext)

	if err == nil {
		// 创建临时文件并复制内容
		if file, err = ioutil.TempFile("/tmp", pattern); err == nil {
			defer readCloser.Close()
			// 将流内容复制到临时文件
			_, err = io.Copy(file, readCloser)
			// 重置文件指针到开始位置
			file.Seek(0, 0)
		}
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
	// 从S3获取对象
	getResponse, err := client.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(client.Config.Bucket),
		Key:    aws.String(client.ToRelativePath(path)),
	})

	return getResponse.Body, err
}

// Put 上传文件到指定路径
// 参数:
//   - urlPath: 文件路径
//   - reader: 文件内容读取器
// 返回:
//   - *oss.Object: 上传成功后的对象信息
//   - error: 错误信息
func (client Client) Put(urlPath string, reader io.Reader) (*oss.Object, error) {
	// 如果reader支持Seek，重置到开始位置
	if seeker, ok := reader.(io.ReadSeeker); ok {
		seeker.Seek(0, 0)
	}

	// 转换为相对路径
	urlPath = client.ToRelativePath(urlPath)
	// 读取所有数据到缓冲区
	buffer, err := ioutil.ReadAll(reader)

	// 检测文件类型
	fileType := mime.TypeByExtension(path.Ext(urlPath))
	if fileType == "" {
		fileType = http.DetectContentType(buffer)
	}

	// 构建上传参数
	params := &s3.PutObjectInput{
		Bucket:        aws.String(client.Config.Bucket), // 存储桶名称（必需）
		Key:           aws.String(urlPath),              // 对象键（必需）
		ACL:           aws.String(client.Config.ACL),    // 访问控制列表
		Body:          bytes.NewReader(buffer),          // 文件内容
		ContentLength: aws.Int64(int64(len(buffer))),    // 内容长度
		ContentType:   aws.String(fileType),             // 内容类型
	}
	// 如果配置了缓存控制，添加到参数中
	if client.Config.CacheControl != "" {
		params.CacheControl = aws.String(client.Config.CacheControl)
	}

	// 执行上传操作
	_, err = client.S3.PutObject(params)

	// 创建返回对象
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
	// 删除S3对象
	_, err := client.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(client.Config.Bucket),
		Key:    aws.String(client.ToRelativePath(path)),
	})
	return err
}

// DeleteObjects 批量删除多个文件
// 参数:
//   - paths: 文件路径列表
// 返回:
//   - error: 错误信息
func (client Client) DeleteObjects(paths []string) (err error) {
	// 构建对象标识符列表
	var objs []*s3.ObjectIdentifier
	for _, v := range paths {
		var obj s3.ObjectIdentifier
		obj.Key = aws.String(strings.TrimPrefix(client.ToRelativePath(v), "/"))
		objs = append(objs, &obj)
	}
	// 构建删除请求参数
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(client.Config.Bucket),
		Delete: &s3.Delete{
			Objects: objs,
		},
	}

	// 执行批量删除操作
	_, err = client.S3.DeleteObjects(input)
	if err != nil {
		return
	}
	return
}

// List 列出指定路径下的所有对象
// 参数:
//   - path: 路径前缀
// 返回:
//   - []*oss.Object: 对象列表
//   - error: 错误信息
func (client Client) List(path string) ([]*oss.Object, error) {
	var objects []*oss.Object
	var prefix string

	// 如果路径不为空，构建前缀
	if path != "" {
		prefix = strings.Trim(path, "/") + "/"
	}

	// 列出S3对象（使用V2版本API）
	listObjectsResponse, err := client.S3.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(client.Config.Bucket),
		Prefix: aws.String(prefix),
	})

	if err == nil {
		// 遍历返回的对象，构建对象列表
		for _, content := range listObjectsResponse.Contents {
			objects = append(objects, &oss.Object{
				Path:             client.ToRelativePath(*content.Key),
				Name:             filepath.Base(*content.Key),
				LastModified:     content.LastModified,
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
		return client.Config.Endpoint
	}

	endpoint := client.S3.Endpoint
	for _, prefix := range []string{"https://", "http://"} {
		endpoint = strings.TrimPrefix(endpoint, prefix)
	}

	return client.Config.Bucket + "." + endpoint
}

var urlRegexp = regexp.MustCompile(`(https?:)?//((\w+).)+(\w+)/`)

// ToRelativePath 将路径转换为相对路径
// 参数:
//   - urlPath: 原始路径
// 返回:
//   - string: 相对路径
func (client Client) ToRelativePath(urlPath string) string {
	if urlRegexp.MatchString(urlPath) {
		if u, err := url.Parse(urlPath); err == nil {
			if client.Config.S3ForcePathStyle { // First part of path will be bucket name
				return strings.TrimPrefix(u.Path, "/"+client.Config.Bucket)
			}
			return u.Path
		}
	}

	if client.Config.S3ForcePathStyle { // First part of path will be bucket name
		return "/" + strings.TrimPrefix(urlPath, "/"+client.Config.Bucket+"/")
	}
	return "/" + strings.TrimPrefix(urlPath, "/")
}

// GetURL 获取文件的公共访问URL
// 参数:
//   - path: 文件路径
// 返回:
//   - string: 公共访问URL
//   - error: 错误信息
func (client Client) GetURL(path string) (url string, err error) {
	if client.Endpoint == "" {
		if client.Config.ACL == s3.BucketCannedACLPrivate || client.Config.ACL == s3.BucketCannedACLAuthenticatedRead {
			getResponse, _ := client.S3.GetObjectRequest(&s3.GetObjectInput{
				Bucket: aws.String(client.Config.Bucket),
				Key:    aws.String(client.ToRelativePath(path)),
			})

			return getResponse.Presign(1 * time.Hour)
		}
	}

	return path, nil
}
