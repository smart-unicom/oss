// Package synology Synology NAS存储服务实现
// 提供Synology NAS的存储接口实现
package synology

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mime/multipart"

	"github.com/smart-unicom/oss"
)

// Client Synology NAS存储客户端
// 封装Synology NAS的操作接口
type Client struct {
	// Config 客户端配置信息
	Config *Config
	// SId 会话ID
	SId string
	// SynoToken Synology令牌
	SynoToken string
	// AppAPIList 应用API列表
	AppAPIList map[string]map[string]interface{}
	// FullAPIList 完整API列表
	FullAPIList map[string]map[string]interface{}
}

// Config Synology NAS客户端配置
// 包含连接Synology NAS所需的所有配置参数
type Config struct {
	// Endpoint 服务端点
	Endpoint string
	// AccessId 访问用户名
	AccessId string
	// AccessKey 访问密码
	AccessKey string
	// SessionExpire 会话是否过期
	SessionExpire bool
	// Verify 是否验证SSL证书
	Verify bool
	// Debug 是否启用调试模式
	Debug bool
	// OtpCode 一次性密码
	OtpCode string
	// SharedFolder 共享文件夹名称
	SharedFolder string
}

// New 初始化Synology NAS存储客户端
// 参数:
//   - config: Synology NAS配置信息
// 返回:
//   - *Client: Synology NAS存储客户端实例
func New(config *Config) *Client {
	// 创建客户端实例
	client := &Client{Config: config}
	// 登录FileStation应用
	client.Login("FileStation")
	// 获取FileStation API列表
	client.GetAPIList("FileStation")
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
	if file, err = ioutil.TempFile("/tmp", "synology"); err == nil {
		defer readCloser.Close()
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
	sharedFolder := client.Config.SharedFolder
	baseURL := client.Config.Endpoint + "/webapi/entry.cgi"
	path = filepath.ToSlash(path)

	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	apiName := "SYNO.FileStation.Download"

	params := url.Values{}
	params.Set("api", apiName)
	params.Set("version", "2")
	params.Set("method", "download")
	params.Set("path", sharedFolder+path)
	params.Set("mode", "download")
	params.Set("SynoToken", client.SynoToken)
	params.Set("_sid", client.SId)

	url := baseURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "stay_login=1; id="+client.SId)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-SYNO-TOKEN", client.SynoToken) // not necessary

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed, status code: %d", resp.StatusCode)
	}

	return resp.Body, err
}

// GetAPIList 获取API列表
// 参数:
//   - app: 应用名称
// 返回:
//   - error: 错误信息
func (client *Client) GetAPIList(app string) error {
	baseURL := client.Config.Endpoint + "/webapi/"
	queryPath := "query.cgi?api=SYNO.API.Info"
	params := url.Values{}
	params.Set("version", "1")
	params.Set("method", "query")
	params.Set("query", "all")

	response, err := http.Get(baseURL + queryPath + "&" + params.Encode())

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return err
	}

	var responseJSON map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&responseJSON)

	if err != nil {
		return err
	}
	responseJSONTwoLevel := make(map[string]map[string]interface{})
	for key, value := range responseJSON["data"].(map[string]interface{}) {
		if innerMap, ok := value.(map[string]interface{}); ok {
			responseJSONTwoLevel[key] = innerMap
		}
	}

	client.AppAPIList = make(map[string]map[string]interface{})
	if app != "" {
		for key := range responseJSONTwoLevel {
			if strings.Contains(strings.ToLower(key), strings.ToLower(app)) {
				client.AppAPIList[key] = responseJSONTwoLevel[key]
			}
		}
	} else {
		client.FullAPIList = responseJSONTwoLevel
	}

	return nil
}

// Login 登录到Synology NAS
// 参数:
//   - application: 应用名称
// 返回:
//   - error: 错误信息
func (client *Client) Login(application string) error {
	baseURL := client.Config.Endpoint + "/webapi/"
	loginAPI := "auth.cgi?api=SYNO.API.Auth"
	params := url.Values{}
	params.Set("version", "3")
	params.Set("method", "login")
	params.Set("account", client.Config.AccessId)
	params.Set("passwd", client.Config.AccessKey)
	params.Set("session", application)
	params.Set("format", "cookie")
	params.Set("enable_syno_token", "yes")

	if client.Config.OtpCode != "" {
		params.Set("opt_code", client.Config.OtpCode)
	}
	loginAPI = loginAPI + "&" + params.Encode()

	var sessionRequestJSON map[string]interface{}
	if !client.Config.SessionExpire && client.SId != "" {
		client.Config.SessionExpire = false
		if client.Config.Debug {
			fmt.Println("User already logged in")
		}
	} else {
		// Check request for error:
		response, err := http.Get(baseURL + loginAPI)
		if err != nil {
			return err
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return err
		}

		err = json.NewDecoder(response.Body).Decode(&sessionRequestJSON)
		if err != nil {
			return err
		}
	}

	// Check DSM response for error:
	errorCode := client.getErrorCode(sessionRequestJSON)

	if errorCode == 0 {
		client.SId = sessionRequestJSON["data"].(map[string]interface{})["sid"].(string)
		client.SynoToken = sessionRequestJSON["data"].(map[string]interface{})["synotoken"].(string)
		client.Config.SessionExpire = false
		if client.Config.Debug {
			fmt.Println("User logged in, new session started!")
		}
	} else {
		client.SId = ""
		if client.Config.Debug {
			fmt.Println("User logged faild")
		}
	}

	return nil

}

// getErrorCode 从响应中获取错误代码
// 参数:
//   - response: API响应数据
// 返回:
//   - int: 错误代码，0表示成功
func (client Client) getErrorCode(response map[string]interface{}) int {
	var code int
	// 检查响应是否成功
	if response["success"].(bool) {
		code = 0 // 无错误
	} else {
		// 提取错误信息
		errorData := response["error"].(map[string]interface{})
		code = int(errorData["code"].(float64))
	}

	return code
}

// Put 上传文件到指定路径
// 参数:
//   - urlPath: 文件上传路径
//   - reader: 文件内容读取器
// 返回:
//   - *oss.Object: 上传成功后的对象信息
//   - error: 错误信息
func (client *Client) Put(urlPath string, reader io.Reader) (r *oss.Object, err error) {
	sharedFolder := client.Config.SharedFolder

	apiName := "SYNO.FileStation.Upload"
	baseURL := client.Config.Endpoint + "/webapi/"
	loginAPI := "entry.cgi"

	params := url.Values{}
	params.Set("api", apiName)
	params.Set("version", "2")
	params.Set("method", "upload")
	params.Set("SynoToken", client.SynoToken)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	parserURL, err := url.Parse(urlPath)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
	}
	path := parserURL.Path
	dir := filepath.Dir(path)
	// change windows path to linux path
	dir = filepath.ToSlash(dir)

	err = writer.WriteField("path", sharedFolder+dir)
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("overwrite", "true")
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("create_parents", "true")
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(urlPath)
	part, err := writer.CreateFormFile("file", filename) // Set a placeholder filename
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, reader)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	url := baseURL + loginAPI + "?" + params.Encode()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "stay_login=1; id="+client.SId)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-SYNO-TOKEN", client.SynoToken) // not necessary

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed, status code: %d", resp.StatusCode)
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
//   - path: 要删除的文件路径
// 返回:
//   - error: 错误信息
func (client Client) Delete(path string) error {
	sharedFolder := client.Config.SharedFolder

	apiName := "SYNO.FileStation.Delete"

	baseURL := client.Config.Endpoint + "/webapi/entry.cgi"
	path = filepath.ToSlash(path)

	params := url.Values{}
	params.Set("api", apiName)
	params.Set("version", "2")
	params.Set("method", "start")
	params.Set("path", sharedFolder+path)
	params.Set("SynoToken", client.SynoToken)
	params.Set("_sid", client.SId)

	req_url := baseURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "stay_login=1; id="+client.SId)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-SYNO-TOKEN", client.SynoToken) // not necessary

	resp, err := http.Get(req_url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return err
	}

	var responseJSON map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseJSON)
	if err != nil {
		return err
	}

	return nil
}

// List 列出指定路径下的所有文件对象
// 参数:
//   - path: 目录路径
// 返回:
//   - []*oss.Object: 文件对象列表
//   - error: 错误信息
func (client Client) List(path string) (objects []*oss.Object, err error) {
	sharedFolder := client.Config.SharedFolder

	apiName := "SYNO.FileStation.List"

	baseURL := client.Config.Endpoint + "/webapi/entry.cgi"
	path = filepath.ToSlash(path)

	params := url.Values{}
	params.Set("api", apiName)
	params.Set("version", "2")
	params.Set("method", "list")
	params.Set("folder_path", sharedFolder+"/"+path)
	params.Set("SynoToken", client.SynoToken)
	params.Set("_sid", client.SId)

	req_url := baseURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "stay_login=1; id="+client.SId)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-SYNO-TOKEN", client.SynoToken) // not necessary

	resp, err := http.Get(req_url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var responseJSON map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseJSON)
	if err != nil {
		return nil, err
	}

	for _, content := range responseJSON["data"].(map[string]interface{})["files"].([]interface{}) {
		now := time.Now()
		path := content.(map[string]interface{})["path"].(string)
		// remove top shared path
		parsedUrl, err := url.Parse(path)
		if err != nil {
			return nil, err
		}
		pathParts := strings.Split(parsedUrl.Path, "/")
		if len(pathParts) > 1 {
			pathParts = append(pathParts[:1], pathParts[2:]...)
		}
		parsedUrl.Path = strings.Join(pathParts, "/")
		path = parsedUrl.String()

		objects = append(objects, &oss.Object{
			Path:             path,
			Name:             filepath.Base(content.(map[string]interface{})["path"].(string)),
			LastModified:     &now,
			StorageInterface: &client,
		})
	}

	return objects, err
}

// GetEndpoint 获取服务端点
// 返回:
//   - string: 服务端点URL
func (client Client) GetEndpoint() string {
	return client.Config.Endpoint
}

// GetURL 获取文件的公共访问URL
// 参数:
//   - path: 文件路径
// 返回:
//   - string: 公共访问URL
//   - error: 错误信息
func (client Client) GetURL(path string) (get_url string, err error) {
	sharedFolder := client.Config.SharedFolder
	baseURL := client.Config.Endpoint + "/webapi/entry.cgi"
	path = filepath.ToSlash(path)

	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	// get file stream
	apiName := "SYNO.FileStation.Download"

	params := url.Values{}
	params.Set("api", apiName)
	params.Set("version", "2")
	params.Set("method", "download")
	params.Set("path", sharedFolder+path)
	params.Set("mode", "download")
	params.Set("SynoToken", client.SynoToken)
	params.Set("_sid", client.SId)

	get_url = baseURL + "?" + params.Encode()

	return get_url, nil
}
