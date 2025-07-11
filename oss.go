// Package oss 对象存储服务抽象层
// 提供统一的对象存储接口，支持多种云存储服务
package oss

import (
	"io"
	"os"
	"time"
)

// StorageInterface 定义对象存储的通用API接口
// 提供文件的上传、下载、删除、列表等基本操作
type StorageInterface interface {
	// Get 获取指定路径的文件
	// 参数:
	//   - path: 文件路径
	// 返回:
	//   - *os.File: 文件对象
	//   - error: 错误信息
	Get(path string) (*os.File, error)
	
	// GetStream 获取指定路径文件的流
	// 参数:
	//   - path: 文件路径
	// 返回:
	//   - io.ReadCloser: 可读流
	//   - error: 错误信息
	GetStream(path string) (io.ReadCloser, error)
	
	// Put 上传文件到指定路径
	// 参数:
	//   - path: 目标路径
	//   - reader: 文件内容读取器
	// 返回:
	//   - *Object: 上传后的对象信息
	//   - error: 错误信息
	Put(path string, reader io.Reader) (*Object, error)
	
	// Delete 删除指定路径的文件
	// 参数:
	//   - path: 文件路径
	// 返回:
	//   - error: 错误信息
	Delete(path string) error
	
	// List 列出指定路径下的所有对象
	// 参数:
	//   - path: 目录路径
	// 返回:
	//   - []*Object: 对象列表
	//   - error: 错误信息
	List(path string) ([]*Object, error)
	
	// GetURL 获取指定路径文件的访问URL
	// 参数:
	//   - path: 文件路径
	// 返回:
	//   - string: 访问URL
	//   - error: 错误信息
	GetURL(path string) (string, error)
	
	// GetEndpoint 获取存储服务的端点地址
	// 返回:
	//   - string: 端点地址
	GetEndpoint() string
}

// Object 存储对象信息
// 包含对象的基本属性和关联的存储接口
type Object struct {
	// Path 对象的完整路径
	Path string
	// Name 对象名称
	Name string
	// LastModified 最后修改时间
	LastModified *time.Time
	// Size 对象大小（字节）
	Size int64
	// StorageInterface 关联的存储接口
	StorageInterface StorageInterface
}

// Get 获取对象的内容
// 返回:
//   - *os.File: 文件对象
//   - error: 错误信息
func (object Object) Get() (*os.File, error) {
	return object.StorageInterface.Get(object.Path)
}
