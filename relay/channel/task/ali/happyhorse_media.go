package ali

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func isHappyHorseTask(req relaycommon.TaskSubmitReq, upstreamModel string) bool {
	return isHappyHorseModel(req.Model) || isHappyHorseModel(upstreamModel)
}

func happyHorseVariantFromModels(req relaycommon.TaskSubmitReq, upstreamModel string) string {
	for _, m := range []string{req.Model, upstreamModel} {
		if v := happyHorseVariantFromModelName(m); v != "" {
			return v
		}
	}
	return ""
}

func happyHorseVariantFromModelName(model string) string {
	lower := strings.ToLower(strings.TrimSpace(model))
	if !strings.Contains(lower, "happyhorse") {
		return ""
	}
	switch {
	case strings.Contains(lower, "video-edit"):
		return "video-edit"
	case strings.Contains(lower, "r2v"):
		return "r2v"
	case strings.Contains(lower, "i2v"):
		return "i2v"
	case strings.Contains(lower, "t2v"):
		return "t2v"
	default:
		return ""
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
	if aliReq != nil {
		add(aliReq.Input.ImgURL)
		add(aliReq.Input.FirstFrameURL)
		for _, item := range aliReq.Input.Media {
			add(item.URL)
		}
	}
	if req.Metadata != nil {
		add(metadataString(req.Metadata, "img_url", "image_url", "first_frame_url"))
		if raw, ok := req.Metadata["reference_images"].([]interface{}); ok {
			for _, item := range raw {
				if s, ok := item.(string); ok {
					add(s)
				}
			}
		}
		if raw, ok := req.Metadata["images"].([]interface{}); ok {
			for _, item := range raw {
				if s, ok := item.(string); ok {
					add(s)
				}
			}
		}
		collectImageURLsFromContent(req.Metadata["content"], add)
		collectImageURLsFromInputObject(req.Metadata["input"], add)
	}
	return urls
}

func collectImageURLsFromContent(raw interface{}, add func(string)) {
	items, ok := raw.([]interface{})
	if !ok {
		return
	}
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		typ := strings.ToLower(metadataString(m, "type"))
		switch typ {
		case "image_url", "image":
			if imageURL, ok := m["image_url"].(map[string]interface{}); ok {
				add(metadataString(imageURL, "url"))
			}
			add(metadataString(m, "url"))
		case "input_image":
			if image, ok := m["image"].(string); ok {
				add(image)
			}
		}
	}
}

func collectImageURLsFromInputObject(raw interface{}, add func(string)) {
	input, ok := raw.(map[string]interface{})
	if !ok {
		return
	}
	if mediaRaw, ok := input["media"].([]interface{}); ok {
		for _, item := range mediaRaw {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			typ := strings.ToLower(metadataString(m, "type"))
			if typ == "first_frame" || typ == "reference_image" || typ == "image" {
				add(metadataString(m, "url"))
			}
		}
	}
	add(metadataString(input, "img_url", "first_frame_url", "image_url"))
}

func applyHappyHorseMedia(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	if !isHappyHorseTask(req, aliReq.Model) {
		return nil
	}
	variant := happyHorseVariantFromModels(req, aliReq.Model)
	if variant == "" && isHappyHorseModel(req.Model) {
		variant = happyHorseVariantFromModelName(req.Model)
	}
	switch variant {
	case "video-edit":
		return applyHappyHorseVideoEdit(req, aliReq)
	case "r2v":
		return applyHappyHorseR2VMedia(req, aliReq)
	case "i2v":
		return applyHappyHorseI2VStyleMedia(req, aliReq, "i2v")
	case "t2v":
		// 文生视频不需要 media；切勿清空 metadata 已合并的 input.media
		return nil
	default:
		return fmt.Errorf("happyhorse: cannot resolve variant (client_model=%q upstream_model=%q)", req.Model, aliReq.Model)
	}
}

func mediaURLsFromAliInput(aliReq *AliVideoRequest) []string {
	if aliReq == nil {
		return nil
	}
	var urls []string
	for _, item := range aliReq.Input.Media {
		if u := strings.TrimSpace(item.URL); u != "" {
			urls = append(urls, u)
		}
	}
	return urls
}

func applyHappyHorseI2VStyleMedia(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest, variant string) error {
	urls := collectImageURLs(req, aliReq)
	if len(urls) == 0 {
		urls = mediaURLsFromAliInput(aliReq)
	}
	if len(urls) == 0 {
		return fmt.Errorf("happyhorse %s requires at least one image (images[], metadata.input.media, or content[].image_url)", variant)
	}
	aliReq.Input.Media = []AliMediaItem{{Type: "first_frame", URL: urls[0]}}
	aliReq.Input.ImgURL = ""
	aliReq.Input.FirstFrameURL = ""
	if variant == "i2v" && aliReq.Parameters != nil {
		aliReq.Parameters.Ratio = ""
	}
	return nil
}

func applyHappyHorseI2V(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	return applyHappyHorseI2VStyleMedia(req, aliReq, "i2v")
}

func applyHappyHorseR2VMedia(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	urls := collectImageURLs(req, aliReq)
	if len(urls) == 0 {
		urls = mediaURLsFromAliInput(aliReq)
	}
	if len(urls) == 0 {
		return fmt.Errorf("happyhorse r2v requires 1-9 reference images")
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

func applyHappyHorseR2V(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	return applyHappyHorseR2VMedia(req, aliReq)
}

func applyHappyHorseVideoEdit(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	videoURL := collectVideoURL(req)
	if videoURL == "" {
		return fmt.Errorf("happyhorse video-edit requires metadata.reference_video_url (public HTTP/HTTPS video URL)")
	}
	if !isPublicHTTPURL(videoURL) {
		return fmt.Errorf("happyhorse video-edit requires a public HTTP(S) video URL")
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
	if aliReq.Parameters != nil {
		aliReq.Parameters.Duration = 0
		if aliReq.Parameters.AudioSetting == "" {
			aliReq.Parameters.AudioSetting = "auto"
		}
	}
	return nil
}

// finalizeHappyHorseAliRequest 欢乐马上游请求唯一出口：模型名 + input.media + 参数。
func finalizeHappyHorseAliRequest(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	if !isHappyHorseTask(req, aliReq.Model) {
		return nil
	}
	if isHappyHorseModel(req.Model) {
		aliReq.Model = strings.TrimSpace(req.Model)
	}
	variant := happyHorseVariantFromModels(req, aliReq.Model)
	if err := applyHappyHorseMedia(req, aliReq); err != nil {
		return err
	}
	applyHappyHorseDefaultVisuals(aliReq, aliReq.Model)
	if variant != "video-edit" && aliReq.Parameters != nil {
		aliReq.Parameters.Duration = clampHappyHorseDuration(aliReq.Parameters.Duration)
	}
	return assertHappyHorseMediaPresent(req, aliReq)
}

// marshalAliVideoRequestBody 欢乐马使用显式 JSON 结构，确保 input.media 必定出现在请求体中。
func marshalAliVideoRequestBody(aliReq *AliVideoRequest, req relaycommon.TaskSubmitReq) ([]byte, error) {
	if !isHappyHorseTask(req, aliReq.Model) {
		return common.Marshal(aliReq)
	}
	if err := assertHappyHorseMediaPresent(req, aliReq); err != nil {
		return nil, err
	}
	payload := map[string]interface{}{
		"model": aliReq.Model,
		"input": map[string]interface{}{
			"prompt": aliReq.Input.Prompt,
			"media":  aliReq.Input.Media,
		},
	}
	if params := happyHorseParametersMap(aliReq, req); len(params) > 0 {
		payload["parameters"] = params
	}
	return json.Marshal(payload)
}

func happyHorseParametersMap(aliReq *AliVideoRequest, req relaycommon.TaskSubmitReq) map[string]interface{} {
	if aliReq.Parameters == nil {
		return nil
	}
	p := aliReq.Parameters
	out := map[string]interface{}{}
	if p.Resolution != "" {
		out["resolution"] = p.Resolution
	}
	// 官方：图生视频不支持 ratio，带上可能导致上游参数异常
	if p.Ratio != "" && !strings.Contains(strings.ToLower(happyHorseVariantFromModels(req, aliReq.Model)), "i2v") {
		out["ratio"] = p.Ratio
	}
	if p.Duration > 0 {
		out["duration"] = p.Duration
	}
	if p.Seed > 0 {
		out["seed"] = p.Seed
	}
	if p.AudioSetting != "" {
		out["audio_setting"] = p.AudioSetting
	}
	if p.Watermark {
		out["watermark"] = p.Watermark
	}
	return out
}

func mergeAliVideoMetadata(metadata map[string]interface{}, aliReq *AliVideoRequest, clientModel string) error {
	if metadata == nil {
		return nil
	}
	if paramsRaw, ok := metadata["parameters"]; ok && paramsRaw != nil {
		paramsBytes, err := json.Marshal(paramsRaw)
		if err != nil {
			return err
		}
		var params AliVideoParameters
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return err
		}
		mergeAliVideoParameters(aliReq.Parameters, &params)
	}
	if inputRaw, ok := metadata["input"]; ok && inputRaw != nil {
		inputBytes, err := json.Marshal(inputRaw)
		if err != nil {
			return err
		}
		var input AliVideoInput
		if err := json.Unmarshal(inputBytes, &input); err != nil {
			return err
		}
		if strings.TrimSpace(input.Prompt) != "" {
			aliReq.Input.Prompt = input.Prompt
		}
		// 仅合并带有效 URL 的 media，避免 metadata 中空 media 覆盖后续组装的值
		if len(input.Media) > 0 && strings.TrimSpace(input.Media[0].URL) != "" {
			aliReq.Input.Media = input.Media
		}
	}
	mergeFlatVideoMetadata(metadata, aliReq, clientModel)
	return nil
}

func assertHappyHorseMediaPresent(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	variant := happyHorseVariantFromModels(req, aliReq.Model)
	switch variant {
	case "i2v", "r2v", "video-edit":
		if len(aliReq.Input.Media) == 0 {
			return fmt.Errorf("happyhorse %s: input.media is required", variant)
		}
		for i, item := range aliReq.Input.Media {
			if strings.TrimSpace(item.Type) == "" || strings.TrimSpace(item.URL) == "" {
				return fmt.Errorf("happyhorse %s: input.media[%d] missing type or url", variant, i)
			}
		}
	}
	return nil
}

func mergeAliVideoParameters(dst *AliVideoParameters, src *AliVideoParameters) {
	if dst == nil || src == nil {
		return
	}
	if src.Resolution != "" {
		dst.Resolution = src.Resolution
		dst.Size = ""
	}
	if src.Size != "" {
		dst.Size = src.Size
	}
	if src.Ratio != "" {
		dst.Ratio = src.Ratio
	}
	if src.Duration > 0 {
		dst.Duration = src.Duration
	}
	if src.Seed > 0 {
		dst.Seed = src.Seed
	}
	if src.AudioSetting != "" {
		dst.AudioSetting = src.AudioSetting
	}
	dst.Watermark = src.Watermark
	if src.PromptExtend {
		dst.PromptExtend = src.PromptExtend
	}
}
