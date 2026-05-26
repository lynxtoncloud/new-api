package tianyi_cache

import (
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel/openai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/gin-gonic/gin"
)

// Adaptor 是 OpenAI Adaptor 的常量注入变体：协议、URL、鉴权、流处理、计费全部继承 openai.Adaptor，
// 仅在 ConvertOpenAIRequest 阶段强制注入两个固定字段到请求体顶层（见 injector.go）。
type Adaptor struct {
	openai.Adaptor
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	out, err := a.Adaptor.ConvertOpenAIRequest(c, info, request)
	if err != nil {
		return nil, err
	}
	injectExtraParams(info)
	return out, nil
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
