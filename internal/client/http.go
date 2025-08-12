package client

import (
	"github.com/go-resty/resty/v2"
	"time"
)

// HTTPClient 是一个封装了 resty.Client 的结构体
type HTTPClient struct {
	client *resty.Client
	token  string
}

// NewHTTPClient 创建一个新的 HTTPClient
func NewHTTPClient(baseURL string, debug bool) *HTTPClient {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetTimeout(30 * time.Second)
	// 使用在 const.go 中定义的默认请求头
	client.SetHeaders(DefaultHeaders)
	client.SetDebug(debug)

	return &HTTPClient{
		client: client,
	}
}

// SetAuthToken 设置并存储认证 token
func (c *HTTPClient) SetAuthToken(token string) {
	c.token = token
}

// R 创建一个新的请求，并自动附加 token
func (c *HTTPClient) R() *resty.Request {
	req := c.client.R()
	if c.token != "" {
		// 根据 curl 请求，将 token 放在名为 "Authorization" 的 header 中
		req.SetHeader("Authorization", c.token)
	}
	return req
}