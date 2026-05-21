package ali

import (
	"encoding/json"
	"strings"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func sumImageURLBytes(req relaycommon.TaskSubmitReq) int {
	n := 0
	for _, u := range req.Images {
		n += len(strings.TrimSpace(u))
	}
	if u := strings.TrimSpace(req.Image); u != "" {
		n += len(u)
	}
	if u := strings.TrimSpace(req.InputReference); u != "" {
		n += len(u)
	}
	return n
}

func mediaTypesFromAliReq(aliReq *AliVideoRequest) []string {
	if aliReq == nil || len(aliReq.Input.Media) == 0 {
		return nil
	}
	out := make([]string, 0, len(aliReq.Input.Media))
	for _, item := range aliReq.Input.Media {
		out = append(out, strings.TrimSpace(item.Type))
	}
	return out
}

func publishHappyHorseRelayDiag(c *gin.Context, taskReq relaycommon.TaskSubmitReq, aliReq *AliVideoRequest, bodyBytes []byte, finalizeErr error) {
	if c == nil || !isHappyHorseTask(taskReq, aliReq.Model) {
		return
	}
	errMsg := ""
	if finalizeErr != nil {
		errMsg = finalizeErr.Error()
	}
	mediaCount := len(aliReq.Input.Media)
	if len(bodyBytes) > 0 {
		if n := peekMarshaledHappyHorseMediaCount(bodyBytes); n >= 0 {
			mediaCount = n
		}
	}
	diag := service.HappyHorseRelayDiag{
		Variant:           happyHorseVariantFromRequest(taskReq, aliReq.Model),
		ClientModel:       strings.TrimSpace(taskReq.Model),
		UpstreamModel:     strings.TrimSpace(aliReq.Model),
		ImagesCount:       len(taskReq.Images),
		ImagesBytes:       sumImageURLBytes(taskReq),
		MediaCount:        mediaCount,
		MediaTypes:        mediaTypesFromAliReq(aliReq),
		UpstreamBodyBytes: len(bodyBytes),
		FinalizeError:     errMsg,
	}
	service.PublishHappyHorseRelayDiagForOpsLog(c, diag)
}

// peekMarshaledHappyHorseMediaCount 从已序列化的上游 JSON 读取 input.media 条数（不记录正文）。
func peekMarshaledHappyHorseMediaCount(bodyBytes []byte) int {
	if len(bodyBytes) == 0 {
		return 0
	}
	var probe struct {
		Input struct {
			Media []json.RawMessage `json:"media"`
		} `json:"input"`
	}
	if err := json.Unmarshal(bodyBytes, &probe); err != nil {
		return -1
	}
	return len(probe.Input.Media)
}
