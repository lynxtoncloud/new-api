package ali

import (
	"fmt"
	"strings"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

// HappyHorse 官方定价：720P 0.9 元/秒，1080P 1.6 元/秒 → 1080/720 倍率。
const happyHorse1080To720Ratio = 160.0 / 90.0

func isHappyHorseModel(model string) bool {
	return strings.Contains(strings.ToLower(model), "happyhorse")
}

// AliMediaItem HappyHorse input.media 元素。
type AliMediaItem struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func normalizeResolutionP(res string) string {
	if res == "" {
		return ""
	}
	r := strings.ToUpper(strings.TrimSpace(res))
	if !strings.HasSuffix(r, "P") {
		r += "P"
	}
	return r
}

func mergeFlatVideoMetadata(metadata map[string]interface{}, aliReq *AliVideoRequest, clientModel string) {
	if metadata == nil || aliReq.Parameters == nil {
		return
	}
	if v, ok := metadata["resolution"].(string); ok && v != "" {
		aliReq.Parameters.Resolution = normalizeResolutionP(v)
		aliReq.Parameters.Size = ""
	}
	if ratio := metadataString(metadata, "ratio", "aspect_ratio", "aspectRatio"); ratio != "" {
		if !strings.Contains(strings.ToLower(clientModel), "happyhorse-1.0-i2v") &&
			!strings.Contains(strings.ToLower(clientModel), "i2v") {
			aliReq.Parameters.Ratio = ratio
		}
	}
	if v, ok := metadata["duration"]; ok {
		if sec, ok := metadataInt(v); ok && sec > 0 {
			aliReq.Parameters.Duration = clampHappyHorseDuration(sec)
		}
	}
	if v, ok := metadata["watermark"].(bool); ok {
		aliReq.Parameters.Watermark = v
	}
	if v, ok := metadata["seed"]; ok {
		if seed, ok := metadataInt(v); ok {
			aliReq.Parameters.Seed = seed
		}
	}
	if v := metadataString(metadata, "audio_setting", "audioSetting"); v != "" {
		aliReq.Parameters.AudioSetting = normalizeHappyHorseAudioSetting(v)
	}
}

func normalizeHappyHorseAudioSetting(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "origin", "keep_origin", "keep-origin", "original":
		return "origin"
	default:
		return "auto"
	}
}

func metadataString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func metadataInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func clampHappyHorseDuration(sec int) int {
	if sec < 3 {
		return 3
	}
	if sec > 15 {
		return 15
	}
	return sec
}

func applyHappyHorseSizeParam(aliReq *AliVideoRequest, size string) error {
	if strings.Contains(size, "*") {
		return fmt.Errorf("happyhorse does not support pixel size %q, use resolution (720P/1080P) and ratio (t2v/r2v)", size)
	}
	aliReq.Parameters.Resolution = normalizeResolutionP(size)
	aliReq.Parameters.Size = ""
	return nil
}

func applyHappyHorseDefaultVisuals(aliReq *AliVideoRequest, model string) {
	if aliReq.Parameters.Resolution == "" && aliReq.Parameters.Size == "" {
		aliReq.Parameters.Resolution = "1080P"
	}
	lower := strings.ToLower(model)
	if strings.Contains(lower, "t2v") && aliReq.Parameters.Ratio == "" {
		aliReq.Parameters.Ratio = "16:9"
	}
	if strings.Contains(lower, "r2v") && aliReq.Parameters.Ratio == "" {
		aliReq.Parameters.Ratio = "16:9"
	}
}

func collectVideoURL(req relaycommon.TaskSubmitReq) string {
	if u := strings.TrimSpace(req.ReferenceVideoURL); u != "" {
		return u
	}
	if len(req.ReferenceVideoURLs) > 0 {
		if u := strings.TrimSpace(req.ReferenceVideoURLs[0]); u != "" {
			return u
		}
	}
	if req.Metadata != nil {
		if v, ok := req.Metadata["reference_video_url"].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
		if raw, ok := req.Metadata["reference_video_urls"].([]interface{}); ok && len(raw) > 0 {
			if s, ok := raw[0].(string); ok {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func isPublicHTTPURL(u string) bool {
	lower := strings.ToLower(strings.TrimSpace(u))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func happyHorseResolutionRatios(model, resolution string) map[string]float64 {
	if !isHappyHorseModel(model) {
		return nil
	}
	switch resolution {
	case "720P":
		return map[string]float64{"resolution-720P": 1}
	case "1080P":
		return map[string]float64{"resolution-1080P": happyHorse1080To720Ratio}
	default:
		return nil
	}
}
