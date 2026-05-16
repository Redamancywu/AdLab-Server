package errors

import (
	"fmt"
	"net/http"
)

// 错误码定义
const (
	// 成功
	CodeSuccess = 0

	// 策略服务 1001-1099
	CodePlacementNotFound  = 1001 // 广告位不存在或 inactive
	CodeStrategyError      = 1002 // 策略服务内部错误

	// S2S 竞价 1101-1199
	CodeS2SBidError        = 1101 // S2S 竞价内部错误
	CodeAllDSPFailed       = 1102 // 所有 DSP 失败
	CodeNoValidBid         = 1103 // 无有效出价

	// C2S 上报 1201-1299
	CodeC2SDuplicateReport = 1201 // 重复上报或参数错误
	CodeC2SReportError     = 1202 // C2S 上报内部错误

	// VAST/追踪 1301-1399
	CodeVASTError          = 1301 // VAST 生成错误
	CodeTrackingError      = 1302 // 追踪事件错误
	CodeMaterialNotFound   = 1303 // 素材不存在

	// 日志查询 1401-1499
	CodeRequestNotFound    = 1401 // 请求 ID 不存在
	CodeLogQueryError      = 1402 // 日志查询错误

	// 管理 API 2001-2099
	CodeValidationFailed   = 2001 // 参数校验失败
	CodeEntityAlreadyExists = 2002 // 实体已存在
	CodeEntityNotFound     = 2003 // 实体不存在
	CodeAdminError         = 2004 // 管理 API 内部错误

	// 系统内部 9000-9999
	CodeDatabaseError      = 9001 // 数据库操作失败
	CodeInternalError      = 9002 // 系统内部错误
)

// AppError 业务错误结构体
type AppError struct {
	Code    int    // 业务错误码
	Message string // 错误消息
	Details string // 详细信息（可选）
	Err     error  // 原始错误（可选）
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 支持 errors.Is/As
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus 返回对应的 HTTP 状态码
func (e *AppError) HTTPStatus() int {
	switch {
	case e.Code == CodeSuccess:
		return http.StatusOK
	case e.Code >= 1001 && e.Code <= 1099:
		return http.StatusNotFound
	case e.Code >= 1101 && e.Code <= 1199:
		return http.StatusOK // S2S 竞价错误通常返回 200 或 204
	case e.Code >= 1201 && e.Code <= 1299:
		return http.StatusBadRequest
	case e.Code >= 1301 && e.Code <= 1399:
		return http.StatusNotFound
	case e.Code >= 1401 && e.Code <= 1499:
		return http.StatusNotFound
	case e.Code >= 2001 && e.Code <= 2099:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// New 创建新的 AppError
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Newf 创建带格式化消息的 AppError
func Newf(code int, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap 包装原始错误
func Wrap(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithDetails 添加详细信息
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// IsNotFound 判断是否为"不存在"类错误
func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == CodePlacementNotFound ||
			appErr.Code == CodeMaterialNotFound ||
			appErr.Code == CodeRequestNotFound ||
			appErr.Code == CodeEntityNotFound
	}
	return false
}

// IsDuplicate 判断是否为"已存在"类错误
func IsDuplicate(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == CodeEntityAlreadyExists ||
			appErr.Code == CodeC2SDuplicateReport
	}
	return false

}
