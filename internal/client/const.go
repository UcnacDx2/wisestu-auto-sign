package client

// DefaultHeaders 存储了除认证信息外的所有默认 HTTP 请求头
var DefaultHeaders = map[string]string{
	"User-Agent":       "Mozilla/5.0 (Linux; Android 15; 22081212C Build/AQ3A.241006.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/138.0.7204.179 Mobile Safari/537.36",
	"App-Version":      "2.0.1",
	"X-Requested-With": "com.neuedu.wisestu",
	"Content-Type":     "application/json",
	"forbid_notify":    "",
}