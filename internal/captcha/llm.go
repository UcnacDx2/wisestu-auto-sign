package captcha

import (
	"encoding/json"
	"errors"

	"github.com/go-resty/resty/v2"
	"zhxg-signin/internal/config"
)

// LLMClient 用于与 LLM API 交互
type LLMClient struct {
	client *resty.Client
	cfg    config.LLMConfig
}

// NewLLMClient 创建一个新的 LLMClient
func NewLLMClient(cfg config.LLMConfig, debug bool) *LLMClient { // <--- 接收 debug 参数
	client := resty.New().
		SetAuthToken(cfg.APIKey).
		SetDebug(debug) // <--- 根据参数设置调试模式
	return &LLMClient{client: client, cfg: cfg}
}

// SolveCaptcha 使用 LLM API 解决验证码
func (c *LLMClient) SolveCaptcha(imageBase64 string) (int, error) {
	prompt := `你是一个精准的图像计算器。你的任务是识别下图中的数学算式并计算出结果。请严格遵循以下步骤和格式：
1. **识别**：准确地识别出图片中的完整数学表达式，忽略任何无关的背景或符号。
2. **计算**：计算出这个表达式的最终数值结果。
3. **输出**：将结果封装在一个JSON对象中，必须包含三个字段：'expression' (识别出的字符串表达式), 'result' (计算出的数字结果), 和 'error' (如果无法识别或计算，则填写错误信息，否则为null)。

**示例**：如果图片内容是 '5 + 3 =', 你应该返回 {"expression": "5+3", "result": 8, "error": null}`

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"model": c.cfg.Model,
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": prompt,
						},
						{
							"type": "image_url",
							"image_url": map[string]string{
								"url": "data:image/png;base64," + imageBase64,
							},
						},
					},
				},
			},
			"max_tokens":  100,
			"temperature": 0.1,
			"stream":      false,
		}).
		Post(c.cfg.Endpoint)

	if err != nil {
		return 0, err
	}

	if resp.IsError() {
		return 0, errors.New("LLM API 请求失败: " + resp.Status())
	}

	var llmResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(resp.Body(), &llmResp); err != nil {
		return 0, err
	}

	if len(llmResp.Choices) == 0 {
		return 0, errors.New("LLM 响应为空")
	}

	var result struct {
		Result int `json:"result"`
	}

	if err := json.Unmarshal([]byte(llmResp.Choices[0].Message.Content), &result); err != nil {
		return 0, err
	}

	return result.Result, nil
}