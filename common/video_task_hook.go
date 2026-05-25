package common

import (
	"context"

	"github.com/gin-gonic/gin"
)

// VideoTaskSuccessInfo 异步视频任务成功时的回调参数（Lynxton 用于 OSS 转存等）。
type VideoTaskSuccessInfo struct {
	TaskID      string
	UserID      int
	Platform    string
	ProviderURL string
}

// OnVideoTaskSucceeded 任务首次标记 SUCCESS 且结算成功后触发；可为 nil。
var OnVideoTaskSucceeded func(ctx context.Context, info VideoTaskSuccessInfo)

// MutateOpenAIVideoJSON 在 GET /v1/videos/:id 返回前改写 JSON（Lynxton 注入 OSS 输出 URL 等）。
var MutateOpenAIVideoJSON func(taskID string, userID int, body []byte) []byte

// StreamVideoOutput 若已从 OSS 转存则直接输出视频流；handled=true 表示已写入响应。
var StreamVideoOutput func(c *gin.Context, taskID string, userID int) (handled bool, err error)
