// Package filesystem 文件系统存储服务实现
// 提供本地文件系统的存储接口实现
package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/smart-unicom/oss"
)

// FileSystem 文件系统存储客户端
// 封装本地文件系统的操作接口
type FileSystem struct {
	// Base 基础目录路径
	Base string
}

// New 初始化文件系统存储客户端
// 参数:
//   - base: 基础目录路径
// 返回:
//   - *FileSystem: 文件系统存储客户端实例
func New(base string) *FileSystem {
	// 获取绝对路径
	absbase, err := filepath.Abs(base)
	if err != nil {
		fmt.Println("FileSystem storage's directory haven't been initialized")
	}
	return &FileSystem{Base: absbase}
}

// GetFullPath 从绝对/相对路径获取完整路径
// 参数:
//   - path: 文件路径
// 返回:
//   - string: 完整路径
func (fileSystem FileSystem) GetFullPath(path string) string {
	fullpath := path
	// 如果不是以基础目录开头，则拼接基础目录
	if !strings.HasPrefix(path, fileSystem.Base) {
		fullpath, _ = filepath.Abs(filepath.Join(fileSystem.Base, path))
	}
	return fullpath
}

// Get 获取指定路径的文件
// 参数:
//   - path: 文件路径
// 返回:
//   - *os.File: 文件对象
//   - error: 错误信息
func (fileSystem FileSystem) Get(path string) (*os.File, error) {
	return os.Open(fileSystem.GetFullPath(path))
}

// GetStream 获取指定路径文件的流
// 参数:
//   - path: 文件路径
// 返回:
//   - io.ReadCloser: 可读流
//   - error: 错误信息
func (fileSystem FileSystem) GetStream(path string) (io.ReadCloser, error) {
	return os.Open(fileSystem.GetFullPath(path))
}

// Put 上传文件到指定路径
// 参数:
//   - path: 目标路径
//   - reader: 文件内容读取器
// 返回:
//   - *oss.Object: 上传后的对象信息
//   - error: 错误信息
func (fileSystem FileSystem) Put(path string, reader io.Reader) (*oss.Object, error) {
	var (
		fullpath = fileSystem.GetFullPath(path)
		// 创建目录结构
		err = os.MkdirAll(filepath.Dir(fullpath), os.ModePerm)
	)

	if err != nil {
		return nil, err
	}

	// 创建目标文件
	dst, err := os.Create(fullpath)

	if err == nil {
		// 如果是可寻址的读取器，重置到开始位置
		if seeker, ok := reader.(io.ReadSeeker); ok {
			seeker.Seek(0, 0)
		}
		// 复制内容到目标文件
		_, err = io.Copy(dst, reader)
	}

	return &oss.Object{Path: path, Name: filepath.Base(path), StorageInterface: fileSystem}, err
}

// Delete 删除指定路径的文件
// 参数:
//   - path: 文件路径
// 返回:
//   - error: 错误信息
func (fileSystem FileSystem) Delete(path string) error {
	return os.Remove(fileSystem.GetFullPath(path))
}

// List 列出指定路径下的所有对象
// 参数:
//   - path: 目录路径
// 返回:
//   - []*oss.Object: 对象列表
//   - error: 错误信息
func (fileSystem FileSystem) List(path string) ([]*oss.Object, error) {
	var (
		objects  []*oss.Object
		fullpath = fileSystem.GetFullPath(path)
	)

	// 遍历目录下的所有文件
	filepath.Walk(fullpath, func(path string, info os.FileInfo, err error) error {
		// 跳过根目录本身
		if path == fullpath {
			return nil
		}

		// 只处理文件，不处理目录
		if err == nil && !info.IsDir() {
			modTime := info.ModTime()
			objects = append(objects, &oss.Object{
				Path:             strings.TrimPrefix(path, fileSystem.Base),
				Name:             info.Name(),
				LastModified:     &modTime,
				StorageInterface: fileSystem,
			})
		}
		return nil
	})

	return objects, nil
}

// GetEndpoint 获取存储服务的端点地址，文件系统的端点是 /
// 返回:
//   - string: 端点地址
func (fileSystem FileSystem) GetEndpoint() string {
	return "/"
}

// GetURL 获取指定路径文件的访问URL
// 参数:
//   - path: 文件路径
// 返回:
//   - string: 访问URL
//   - error: 错误信息
func (fileSystem FileSystem) GetURL(path string) (url string, err error) {
	return path, nil
}
