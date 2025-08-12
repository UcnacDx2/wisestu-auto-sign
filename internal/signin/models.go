package signin

// BaseResponse 是许多 API 响应的基础结构
type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// VerificationResponse 验证码请求的响应
type VerificationResponse struct {
	VerificationID    string `json:"verification_id"`
	VerificationImage string `json:"verification_image"`
}

// LoginRequest 登录请求的结构
type LoginRequest struct {
	Action             string `json:"action"`
	VerificationID     string `json:"verification_id"`
	VerificationImage  string `json:"verification_image"`
	VerificationAnswer string `json:"verification_answer"`
	LoginName          string `json:"login_name"`
	Password           string `json:"password"`
	ClientType         string `json:"client_type"`
	ClientVer          string `json:"client_ver"`
	ClientExtra        string `json:"client_extra"`
}

// UnSigninListResponse 未签到列表的响应
type UnSigninListResponse struct {
	Result struct {
		List []struct {
			ID             int    `json:"id"`
			SigninTypeName string `json:"signin_type_name"`
		} `json:"list"`
	} `json:"result"`
}

// CheckOutsideFlagRequest 点击签到请求的结构
type CheckOutsideFlagRequest struct {
	Action string  `json:"action"`
	ID     int     `json:"id"`
	Lng    float64 `json:"lng"`
	Lat    float64 `json:"lat"`
}