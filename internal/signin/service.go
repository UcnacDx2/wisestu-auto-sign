package signin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"zhxg-signin/internal/captcha"
	"zhxg-signin/internal/client"
	"zhxg-signin/internal/config"
	"zhxg-signin/internal/logger"
)

// Service 封装了签到服务的所有逻辑
type Service struct {
	cfg        config.Config
	httpClient *client.HTTPClient
	llmClient  *captcha.LLMClient
	log        *zap.Logger
	token      string
}

// NewService 创建一个新的签到服务
func NewService(cfg config.Config) *Service {
	return &Service{
		cfg:        cfg,
		httpClient: client.NewHTTPClient(cfg.SignIn.BaseURL, cfg.Logging.Debug),
		llmClient:  captcha.NewLLMClient(cfg.LLM, cfg.Logging.Debug),
		log:        logger.GetLogger(),
	}
}

// Run 执行完整的签到流程
func (s *Service) Run() error {
	s.log.Info("开始签到流程")

	// 阶段一：检查登录状态
	loggedIn, err := s.checkLoginStatus()
	if err != nil {
		s.log.Warn("检查登录状态时出错，将尝试重新登录", zap.Error(err))
	}

	if loggedIn {
		s.log.Info("Token 有效，已处于登录状态")
	} else {
		s.log.Info("Token 无效或不存在，需要登录")
		// 阶段二：执行登录循环
		token, err := s.login()
		if err != nil {
			s.log.Error("登录流程失败", zap.Error(err))
			return err
		}
		s.token = token
		s.log.Info("登录成功，获取到新的 Token")
	}

	s.httpClient.SetAuthToken(s.token)

	// 阶段三：执行签到
	return s.performSignInFlow()
}

// checkLoginStatus 检查当前 token 是否有效
func (s *Service) checkLoginStatus() (bool, error) {
	// 即使 token 为空，也尝试请求，让服务器决定状态
	// s.httpClient.R() 会自动附加 s.token (如果存在)
	resp, err := s.httpClient.R().
		SetBody(`{"action":"queryMyStuInfo"}`).
		Post("/dnui/api/student/basic/stuInfo.api")

	if err != nil {
		s.log.Warn("检查登录状态请求失败", zap.Error(err))
		return false, err
	}

	var baseResp BaseResponse
	if err := json.Unmarshal(resp.Body(), &baseResp); err != nil {
		s.log.Warn("解析登录状态响应失败", zap.Error(err))
		return false, fmt.Errorf("解析登录状态响应失败: %w", err)
	}

	s.log.Info("检查登录状态响应", zap.Int("code", baseResp.Code), zap.String("message", baseResp.Message))
	// code 为 0 表示已登录
	return baseResp.Code == 0, nil
}

// login 执行带重试的登录循环
func (s *Service) login() (string, error) {
	s.token = "" // 循环开始前清除 token

	var lastErr error
	for i := 0; i < 5; i++ {
		s.log.Info("开始登录尝试", zap.Int("attempt", i+1))

		// 1. 获取验证码
		verifResp, err := s.getVerification()
		if err != nil {
			lastErr = fmt.Errorf("第 %d 次尝试：获取验证码失败: %w", i+1, err)
			s.log.Warn(lastErr.Error())
			time.Sleep(s.cfg.SignIn.RetryInterval)
			continue
		}

		// 2. 识别验证码
		answer, err := s.llmClient.SolveCaptcha(verifResp.VerificationImage)
		if err != nil {
			lastErr = fmt.Errorf("第 %d 次尝试：识别验证码失败: %w", i+1, err)
			s.log.Warn(lastErr.Error())
			time.Sleep(s.cfg.SignIn.RetryInterval)
			continue
		}
		s.log.Info("验证码识别结果", zap.Int("answer", answer))

		// 3. 尝试登录
		reqBody := LoginRequest{
			Action:             "loginStudent",
			VerificationID:     verifResp.VerificationID,
			VerificationImage:  verifResp.VerificationImage,
			VerificationAnswer: strconv.Itoa(answer),
			LoginName:          s.cfg.User.Username,
			Password:           s.cfg.User.Password,
			ClientType:         "App",
			ClientVer:          "2.0.1",
			ClientExtra:        `{"available":true,"platform":"Android","version":"15","uuid":"","cordova":"8.1.0","model":"22081212C","manufacturer":"Xiaomi","isVirtual":false,"serial":"unknown"}`,
		}

		resp, err := s.httpClient.R().
			SetBody(reqBody).
			Post("/dnui/api/user/loginout.api")

		if err != nil {
			lastErr = fmt.Errorf("第 %d 次尝试：登录请求失败: %w", i+1, err)
			s.log.Warn(lastErr.Error())
			time.Sleep(s.cfg.SignIn.RetryInterval)
			continue
		}

		var baseResp BaseResponse
		if err := json.Unmarshal(resp.Body(), &baseResp); err != nil {
			lastErr = fmt.Errorf("第 %d 次尝试：解析登录响应失败: %w", i+1, err)
			s.log.Warn(lastErr.Error())
			time.Sleep(s.cfg.SignIn.RetryInterval)
			continue
		}

		s.log.Info("登录响应", zap.Int("code", baseResp.Code), zap.String("message", baseResp.Message))

		// 4. 判定循环退出条件
		if baseResp.Code == 0 {
			token := resp.Header().Get("token")
			if token == "" {
				lastErr = errors.New("登录成功但未在响应头中找到 token")
				s.log.Warn(lastErr.Error())
				time.Sleep(s.cfg.SignIn.RetryInterval)
				continue // 虽然 code 为 0，但没 token 还是得重试
			}
			return token, nil // 成功获取 token，退出循环
		}

		if baseResp.Code == 1002 {
			return "", fmt.Errorf("登录失败：密码错误 (code: 1002)")
		}

		lastErr = fmt.Errorf("第 %d 次尝试：登录失败，code: %d, message: %s", i+1, baseResp.Code, baseResp.Message)
		s.log.Warn(lastErr.Error())
		time.Sleep(s.cfg.SignIn.RetryInterval)
	}

	return "", fmt.Errorf("登录失败，已达到最大重试次数 (5次): %w", lastErr)
}

func (s *Service) getVerification() (*VerificationResponse, error) {
	var result VerificationResponse
	resp, err := s.httpClient.R().
		SetBody(map[string]string{"action": "queryVerificationQuestion", "client_type": "App"}).
		SetResult(&result).
		Post("/dnui/api/user/loginout.api")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("获取验证码请求失败: %s", resp.Status())
	}
	return &result, nil
}

func (s *Service) performSignInFlow() error {
	// 获取未签到列表
	var listResp UnSigninListResponse
	resp, err := s.httpClient.R().
		SetBody(map[string]interface{}{"action": "getUnSigninList", "pageSize": 10, "pageNum": 1}).
		SetResult(&listResp).
		Post("/dnui/api/student/signin/signin.api")

	if err != nil {
		return fmt.Errorf("获取签到列表请求失败: %w", err)
	}

	var baseResp BaseResponse
	if err := json.Unmarshal(resp.Body(), &baseResp); err != nil {
		return fmt.Errorf("解析签到列表响应失败: %w", err)
	}

	if baseResp.Code != 0 {
		return fmt.Errorf("获取签到列表失败, code: %d, message: %s", baseResp.Code, baseResp.Message)
	}

	if len(listResp.Result.List) == 0 {
		s.log.Info("没有需要签到的任务")
		return nil
	}

	var signinID int
	var batchNo int
	foundInternship := false

	// 遍历列表查找 "实习" 签到任务
	for _, item := range listResp.Result.List {
		if item.SigninTypeName == "实习" {
			signinID = item.ID
			batchNo = item.BatchNo
			foundInternship = true
			s.log.Info("找到 '实习' 签到任务", zap.Int("signinID", signinID), zap.Int("batchNo", batchNo))
			break
		}
	}

	if !foundInternship {
		s.log.Warn("未找到 '实习' 签到任务，将使用列表中的第一个任务进行签到")
		signinID = listResp.Result.List[0].ID
		batchNo = listResp.Result.List[0].BatchNo // 假设第一个任务也有 batch_no
		s.log.Info("使用第一个签到任务", zap.Int("signinID", signinID), zap.Int("batchNo", batchNo))
	}

	// 1. 调用“进入签到”接口
	if err := s.getSigninDetails(signinID, batchNo); err != nil {
		return fmt.Errorf("进入签到失败: %w", err)
	}

	// 2. 调用“updateLocationSignin”接口
	if err := s.updateLocationSignin(signinID, batchNo); err != nil {
		return fmt.Errorf("提交位置签到失败: %w", err)
	}

	// 3. 调用“签到情况”接口
	if err := s.getSigninSuccess(signinID, batchNo); err != nil {
		return fmt.Errorf("获取签到情况失败: %w", err)
	}

	s.log.Info("签到流程执行完毕")
	return nil
}

// getSigninDetails 调用“进入签到”接口
func (s *Service) getSigninDetails(signinID, batchNo int) error {
	reqBody := GetSigninDetailsRequest{
		Action:  "getSigninDetails",
		ID:      signinID,
		BatchNo: batchNo,
	}

	resp, err := s.httpClient.R().
		SetBody(reqBody).
		Post("/dnui/api/student/signin/signin.api")

	if err != nil {
		return fmt.Errorf("进入签到请求失败: %w", err)
	}

	var baseResp BaseResponse
	if err := json.Unmarshal(resp.Body(), &baseResp); err != nil {
		return fmt.Errorf("解析进入签到响应失败: %w", err)
	}

	if baseResp.Code != 0 {
		return fmt.Errorf("进入签到失败, code: %d, message: %s", baseResp.Code, baseResp.Message)
	}

	s.log.Info("成功进入签到", zap.Int("signinID", signinID), zap.Int("batchNo", batchNo))
	return nil
}

// updateLocationSignin 调用“updateLocationSignin”接口
func (s *Service) updateLocationSignin(signinID, batchNo int) error {
	// 构建 signin_location 字段的 JSON 字符串
	signinLocation := SigninLocation{
		Point: SigninLocationPoint{
			Lng: s.cfg.Location.Longitude,
			Lat: s.cfg.Location.Latitude,
		},
		Address: "柳州市鱼峰区葡萄山路7号科技楼", // 从 Apifox CLI 中获取的固定地址
		AddressComponents: SigninLocationAddressComponents{
			StreetNumber: "7号",
			Street:       "葡萄山路",
			District:     "鱼峰区",
			City:         "柳州市",
			Province:     "广西壮族自治区",
		},
	}

	signinLocationJSON, err := json.Marshal(signinLocation)
	if err != nil {
		return fmt.Errorf("序列化 signin_location 失败: %w", err)
	}

	reqBody := UpdateLocationSigninRequest{
		Action:         "updateLocationSignin",
		ID:             signinID,
		BatchNo:        batchNo,
		SigninLocation: string(signinLocationJSON),
		OutsideFlag:    "1", // 从 Apifox CLI 中获取的固定值
	}

	resp, err := s.httpClient.R().
		SetBody(reqBody).
		Post("/dnui/api/student/signin/signin.api")

	if err != nil {
		return fmt.Errorf("提交位置签到请求失败: %w", err)
	}

	var baseResp BaseResponse
	if err := json.Unmarshal(resp.Body(), &baseResp); err != nil {
		return fmt.Errorf("解析位置签到响应失败: %w", err)
	}

	if baseResp.Code != 0 {
		return fmt.Errorf("位置签到失败, code: %d, message: %s", baseResp.Code, baseResp.Message)
	}

	s.log.Info("位置签到成功", zap.Int("signinID", signinID), zap.Int("batchNo", batchNo))
	return nil
}

// getSigninSuccess 调用“签到情况”接口
func (s *Service) getSigninSuccess(signinID, batchNo int) error {
	reqBody := GetSigninSuccessRequest{
		Action:  "getSigninSuccess",
		ID:      signinID,
		BatchNo: batchNo,
	}

	resp, err := s.httpClient.R().
		SetBody(reqBody).
		Post("/dnui/api/student/signin/signin.api")

	if err != nil {
		return fmt.Errorf("获取签到情况请求失败: %w", err)
	}

	var baseResp BaseResponse
	if err := json.Unmarshal(resp.Body(), &baseResp); err != nil {
		return fmt.Errorf("解析签到情况响应失败: %w", err)
	}

	if baseResp.Code != 0 {
		return fmt.Errorf("获取签到情况失败, code: %d, message: %s", baseResp.Code, baseResp.Message)
	}

	s.log.Info("成功获取签到情况", zap.Int("signinID", signinID), zap.Int("batchNo", batchNo))
	return nil
}