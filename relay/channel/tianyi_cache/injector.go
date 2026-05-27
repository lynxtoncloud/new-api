package tianyi_cache

import (
	"encoding/json"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

const ChannelName = "tianyi_cache"

// 两个强制注入的常量参数。后续要改的话，只改这里。
const (
	paramKeyUser    = "user"
	paramValueUser  = "32a25556d3c5b0b5f1b1e8c9612996f6"
	paramKeyCaching = "caching"
)

// paramValueCaching 是注入到请求体顶层的 caching 字段值: {"type": "enabled"}
var paramValueCaching = json.RawMessage(`{"type":"enabled"}`)

// injectExtraParams 把两个常量字段写入 info.ParamOverride。
// 上游 compatible_handler 在 marshal 后会调用 ApplyParamOverrideWithRelayInfo，
// 将 ParamOverride 的键值对 patch 到请求 JSON 顶层，从而完成注入。
// 选择 ParamOverride 路径而不是修改 request 自身，是为了：
//  1. 保留 *dto.GeneralOpenAIRequest 类型，不破坏上游对 type assertion 的依赖（如 SystemPrompt 注入）；
//  2. caching 字段不在 GeneralOpenAIRequest struct 上，需要 marshal 后才能 patch。
//
// 同名 key 会被强制覆盖：本变体渠道的存在意义就是固定这两个值。
func injectExtraParams(info *relaycommon.RelayInfo) {
	if info == nil || info.ChannelMeta == nil {
		return
	}
	if info.ParamOverride == nil {
		info.ParamOverride = map[string]interface{}{}
	}
	info.ParamOverride[paramKeyUser] = paramValueUser
	info.ParamOverride[paramKeyCaching] = paramValueCaching
}
