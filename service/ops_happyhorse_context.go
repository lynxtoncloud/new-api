package service

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/gin-gonic/gin"
)

const opsHappyHorseSnapshotFrom = "new-api"

// HappyHorseRelayDiag 欢乐马转发诊断（仅关键字段，不含图片/视频 URL 或 base64 正文）。
type HappyHorseRelayDiag struct {
	Variant           string
	ClientModel       string
	UpstreamModel     string
	ImagesCount       int
	ImagesBytes       int
	MediaCount        int
	MediaTypes        []string
	UpstreamBodyBytes int
	FinalizeError     string
}

// PublishHappyHorseRelayDiagForOpsLog 写入 Gin context，供 Lynxton relay_request_summary 以 Zap JSON 字段输出。
func PublishHappyHorseRelayDiagForOpsLog(c *gin.Context, d HappyHorseRelayDiag) {
	if c == nil {
		return
	}
	common.SetContextKey(c, constant.ContextKeyOpsHappyHorseSnapshotFrom, opsHappyHorseSnapshotFrom)
	common.SetContextKey(c, constant.ContextKeyOpsHappyHorseVariant, strings.TrimSpace(d.Variant))
	common.SetContextKey(c, constant.ContextKeyOpsHappyHorseClientModel, strings.TrimSpace(d.ClientModel))
	common.SetContextKey(c, constant.ContextKeyOpsHappyHorseUpstreamModel, strings.TrimSpace(d.UpstreamModel))
	common.SetContextKey(c, constant.ContextKeyOpsHappyHorseImagesCount, d.ImagesCount)
	common.SetContextKey(c, constant.ContextKeyOpsHappyHorseImagesBytes, d.ImagesBytes)
	common.SetContextKey(c, constant.ContextKeyOpsHappyHorseMediaCount, d.MediaCount)
	common.SetContextKey(c, constant.ContextKeyOpsHappyHorseMediaTypes, strings.Join(d.MediaTypes, ","))
	common.SetContextKey(c, constant.ContextKeyOpsHappyHorseUpstreamBodyBytes, d.UpstreamBodyBytes)
	common.SetContextKey(c, constant.ContextKeyOpsHappyHorseFinalizeError, strings.TrimSpace(d.FinalizeError))
}
