// Package azureblob 提供Azure Blob存储的实现
// 支持Azure Blob存储服务的文件上传、下载、删除等操作
package azureblob

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

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/smart-unicom/oss"
)

// Client Azure Blob存储客户端
// 封装了Azure Blob存储的操作接口
type Client struct {
	Config       *Config                // 配置信息
	containerURL *azblob.ContainerURL   // 容器URL对象
}

// Config Azure Blob存储配置
// 包含连接Azure Blob存储所需的所有配置信息
type Config struct {
	AccessId  string // 账户名称
	AccessKey string // 访问密钥
	Region    string // 区域
	Bucket    string // 容器名称
	Endpoint  string // 端点URL
}

// urlRegexp URL正则表达式，用于匹配HTTP/HTTPS URL格式
var urlRegexp = regexp.MustCompile(`(https?:)?//((\w+).)+(\w+)/`)

// ToRelativePath 将路径转换为相对路径
// 参数:
//   - urlPath: 原始URL路径
// 返回:
//   - string: 处理后的相对路径
func (client Client) ToRelativePath(urlPath string) string {
	// 如果是完整URL，解析并提取路径部分
	if urlRegexp.MatchString(urlPath) {
		if u, err := url.Parse(urlPath); err == nil {
			return strings.TrimPrefix(u.Path, "/")
		}
	}

	// 移除路径前缀的斜杠
	return strings.TrimPrefix(urlPath, "/")
}

// blobFormatString Azure Blob存储的URL格式模板
const blobFormatString = `https://%s.blob.core.windows.net`

var (
	// ctx 全局上下文，用于Azure Blob操作
	ctx = context.Background()
)

// New 创建新的Azure Blob存储客户端
// 参数:
//   - config: Azure Blob存储配置
// 返回:
//   - *Client: Azure Blob存储客户端实例
func New(config *Config) *Client {
	// 创建客户端实例
	var client = &Client{Config: config}

	// 获取服务URL并初始化容器URL
	serviceURL, _ := GetBlobService(config)
	client.containerURL = containerUrl(serviceURL, config)
	return client
}

// GetBlobService 获取Azure Blob服务URL
// 参数:
//   - config: Azure Blob存储配置
// 返回:
//   - azblob.ServiceURL: 服务URL对象
//   - error: 错误信息
func GetBlobService(config *Config) (azblob.ServiceURL, error) {
	// 使用存储账户名称和密钥创建凭据对象
	credential, err := azblob.NewSharedKeyCredential(config.AccessId, config.AccessKey)
	if err != nil {
		return azblob.ServiceURL{}, err
	}

	// 创建请求管道，用于处理HTTP(S)请求和响应
	// 在更高级的场景中，可以配置遥测、重试策略、日志记录等选项
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// 从Azure门户获取存储账户的Blob服务URL端点
	// URL通常格式为: https://accountname.blob.core.windows.net
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, config.AccessId))

	// 创建包装服务URL和请求管道的ServiceURL对象
	return azblob.NewServiceURL(*u, p), nil
}

// containerUrl 获取容器URL对象
// 参数:
//   - serviceURL: 服务URL对象
//   - config: Azure Blob存储配置
// 返回:
//   - *azblob.ContainerURL: 容器URL对象指针
func containerUrl(serviceURL azblob.ServiceURL, config *Config) *azblob.ContainerURL {
	// 返回包装容器URL和请求管道的ContainerURL对象
	container := serviceURL.NewContainerURL(config.Bucket)
	return &container
}

// UploadBlob 上传Blob到Azure存储
// 参数:
//   - blobName: Blob名称
//   - blobType: Blob内容类型
//   - data: 要上传的数据流
// 返回:
//   - azblob.BlockBlobURL: 块Blob URL对象
//   - error: 错误信息
func (client Client) UploadBlob(blobName *string, blobType *string, data io.ReadSeeker) (azblob.BlockBlobURL, error) {
	// 创建引用Azure存储账户容器中Blob的URL
	// 返回包装Blob URL和请求管道的BlockBlobURL对象
	blobURL := client.containerURL.NewBlockBlobURL(*blobName) // Blob名称可以是混合大小写

	// 上传Blob数据
	_, err := blobURL.Upload(ctx, data, azblob.BlobHTTPHeaders{ContentType: *blobType}, azblob.Metadata{}, azblob.BlobAccessConditions{}, azblob.DefaultAccessTier, nil, azblob.ClientProvidedKeyOptions{}, azblob.ImmutabilityPolicyOptions{})
	if err != nil {
		return azblob.BlockBlobURL{}, err
	}

	return blobURL, nil
}

// DownloadBlob 从Azure存储下载Blob
// 参数:
//   - blobName: Blob名称
// 返回:
//   - *azblob.DownloadResponse: 下载响应对象
//   - error: 错误信息
func (client Client) DownloadBlob(blobName *string) (*azblob.DownloadResponse, error) {
	// 创建引用Azure存储账户容器中Blob的URL
	// 返回包装Blob URL和请求管道的BlockBlobURL对象
	blobURL := client.containerURL.NewBlockBlobURL(*blobName) // Blob名称可以是混合大小写

	// 下载Blob内容并验证操作是否成功
	return blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
}

// DeleteBlob 从Azure存储删除Blob
// 参数:
//   - blobName: Blob名称
// 返回:
//   - error: 错误信息
func (client Client) DeleteBlob(blobName *string) error {
	// 创建引用Azure存储账户容器中Blob的URL
	// 返回包装Blob URL和请求管道的BlockBlobURL对象
	blobURL := client.containerURL.NewBlockBlobURL(*blobName) // Blob名称可以是混合大小写

	// 删除Blob
	_, err := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	if err != nil {
		return err
	}

	return nil
}

// GetListBlob 获取容器中的Blob列表
// 返回:
//   - [][]azblob.BlobItemInternal: Blob项目的二维数组
//   - error: 错误信息
func (client Client) GetListBlob() ([][]azblob.BlobItemInternal, error) {
	var results [][]azblob.BlobItemInternal

	// 列出容器中的Blob；由于容器可能包含数百万个Blob，因此分段进行
	for marker := (azblob.Marker{}); marker.NotDone(); { // Marker{}周围的括号是必需的，以避免编译器错误
		// 获取从当前标记指示的Blob开始的结果段
		listBlob, err := client.containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return nil, err
		}
		// 重要：ListBlobs返回下一段的开始；必须使用此标记获取下一段
		marker = listBlob.NextMarker

		// 处理此结果段中返回的Blob（如果段为空，则循环体不会执行）
		for _, blobInfo := range listBlob.Segment.BlobItems {
			fmt.Print("Blob name: " + blobInfo.Name + "\n")
		}

		results = append(results, listBlob.Segment.BlobItems)
	}

	return results, nil
}

// Get 获取指定路径的文件
// 参数:
//   - path: 文件路径
// 返回:
//   - *os.File: 文件对象
//   - error: 错误信息
func (client Client) Get(path string) (file *os.File, err error) {
	// 转换为相对路径
	path = client.ToRelativePath(path)
	// 获取文件流
	readCloser, err := client.GetStream(path)

	if err == nil {
		// 创建临时文件
		if file, err = ioutil.TempFile("/tmp", "ali"); err == nil {
			defer func(readCloser io.ReadCloser) {
				err := readCloser.Close()
				if err != nil {

				}
			}(readCloser)
			// 将流内容复制到临时文件
			_, err = io.Copy(file, readCloser)
			// 重置文件指针到开始位置
			_, err := file.Seek(0, 0)
			if err != nil {
				return nil, err
			}
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
	name := path
	// 下载Blob并返回响应体
	blob, err := client.DownloadBlob(&name)
	if err != nil {
		return nil, err
	}
	return blob.Response().Body, err
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
		_, err := seeker.Seek(0, 0)
		if err != nil {
			return nil, err
		}
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

	if fileType == "" {
		fileType = http.DetectContentType(buffer)
	}

	// 上传Blob到Azure存储
	_, err = client.UploadBlob(&urlPath, &fileType, bytes.NewReader(buffer))
	if err != nil {
		return nil, err
	}
	now := time.Now()

	// 创建返回对象
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
	// 转换为相对路径
	path = client.ToRelativePath(path)
	return client.DeleteBlob(&path)
}

// List 列出指定路径下的所有对象
// 参数:
//   - path: 路径前缀
// 返回:
//   - []*oss.Object: 对象列表
//   - error: 错误信息
func (client Client) List(path string) ([]*oss.Object, error) {
	panic("implement me")
}

// GetURL 获取文件的访问URL
// 参数:
//   - path: 文件路径
// 返回:
//   - string: 访问URL
//   - error: 错误信息
func (client Client) GetURL(path string) (string, error) {
	return path, nil
}

// GetEndpoint 获取存储端点
// 返回:
//   - string: 存储端点URL
func (client Client) GetEndpoint() string {
	// 如果配置了自定义端点，使用自定义端点
	if client.Config.Endpoint != "" {
		return client.Config.Endpoint
	}
	// 否则使用默认的Azure Blob存储端点格式
	return fmt.Sprintf(blobFormatString, client.Config.AccessId)
}
