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

func mergeFlatVideoMetadata(metadata map[string]interface{}, aliReq *AliVideoRequest) {
	if metadata == nil || aliReq.Parameters == nil {
		return
	}
	if v, ok := metadata["resolution"].(string); ok && v != "" {
		aliReq.Parameters.Resolution = normalizeResolutionP(v)
		aliReq.Parameters.Size = ""
	}
	if ratio := metadataString(metadata, "ratio", "aspect_ratio", "aspectRatio"); ratio != "" {
		aliReq.Parameters.Ratio = ratio
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

func collectImageURLs(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) []string {
	seen := map[string]bool{}
	var urls []string
	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			return
		}
		seen[u] = true
		urls = append(urls, u)
	}
	for _, img := range req.Images {
		add(img)
	}
	add(req.InputReference)
	add(req.Image)
	add(aliReq.Input.ImgURL)
	add(aliReq.Input.FirstFrameURL)
	if req.Metadata != nil {
		if v, ok := req.Metadata["img_url"].(string); ok {
			add(v)
		}
		if v, ok := req.Metadata["first_frame_url"].(string); ok {
			add(v)
		}
		if raw, ok := req.Metadata["reference_images"].([]interface{}); ok {
			for _, item := range raw {
				if s, ok := item.(string); ok {
					add(s)
				}
			}
		}
	}
	return urls
}

func collectVideoURL(req relaycommon.TaskSubmitReq) string {
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

// applyHappyHorseMedia 按官方 API 组装 input.media（t2v 无需 media）。
func applyHappyHorseMedia(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	if !isHappyHorseModel(aliReq.Model) {
		return nil
	}
	lower := strings.ToLower(aliReq.Model)
	switch {
	case strings.Contains(lower, "video-edit"):
		return applyHappyHorseVideoEdit(req, aliReq)
	case strings.Contains(lower, "r2v"):
		return applyHappyHorseR2V(req, aliReq)
	case strings.Contains(lower, "i2v"):
		return applyHappyHorseI2V(req, aliReq)
	default:
		// t2v：仅需 prompt + parameters
		return nil
	}
}

func applyHappyHorseI2V(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	urls := collectImageURLs(req, aliReq)
	if len(urls) == 0 {
		return fmt.Errorf("happyhorse i2v requires one first-frame image (images[] or metadata.img_url)")
	}
	aliReq.Input.Media = []AliMediaItem{{Type: "first_frame", URL: urls[0]}}
	aliReq.Input.ImgURL = ""
	aliReq.Input.FirstFrameURL = ""
	return nil
}

func applyHappyHorseR2V(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	urls := collectImageURLs(req, aliReq)
	if len(urls) == 0 {
		return fmt.Errorf("happyhorse r2v requires 1-9 reference images in images[]")
	}
	if len(urls) > 9 {
		urls = urls[:9]
	}
	media := make([]AliMediaItem, 0, len(urls))
	for _, u := range urls {
		media = append(media, AliMediaItem{Type: "reference_image", URL: u})
	}
	aliReq.Input.Media = media
	aliReq.Input.ImgURL = ""
	aliReq.Input.FirstFrameURL = ""
	return nil
}

func applyHappyHorseVideoEdit(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	videoURL := collectVideoURL(req)
	if videoURL == "" {
		return fmt.Errorf("happyhorse video-edit requires metadata.reference_video_url (public HTTP/HTTPS video URL)")
	}
	if !isPublicHTTPURL(videoURL) {
		return fmt.Errorf("happyhorse video-edit requires a public HTTP(S) video URL; data URLs or local files are not supported by upstream")
	}
	media := []AliMediaItem{{Type: "video", URL: videoURL}}
	refImages := collectImageURLs(req, aliReq)
	if len(refImages) > 5 {
		refImages = refImages[:5]
	}
	for _, u := range refImages {
		media = append(media, AliMediaItem{Type: "reference_image", URL: u})
	}
	aliReq.Input.Media = media
	aliReq.Input.ImgURL = ""
	aliReq.Input.FirstFrameURL = ""
	// 视频编辑时长与输入对齐，不传 duration
	aliReq.Parameters.Duration = 0
	if aliReq.Parameters.AudioSetting == "" {
		aliReq.Parameters.AudioSetting = "auto"
	}
	return nil
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
