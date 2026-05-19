package ali

import (
	"encoding/json"
	"fmt"
	"strings"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func isHappyHorseTask(req relaycommon.TaskSubmitReq, upstreamModel string) bool {
	return isHappyHorseModel(req.Model) || isHappyHorseModel(upstreamModel)
}

func happyHorseVariantFromModels(req relaycommon.TaskSubmitReq, upstreamModel string) string {
	for _, m := range []string{req.Model, upstreamModel} {
		lower := strings.ToLower(strings.TrimSpace(m))
		if !strings.Contains(lower, "happyhorse") {
			continue
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
		}
	}
	return ""
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
	for _, item := range aliReq.Input.Media {
		add(item.URL)
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
	switch happyHorseVariantFromModels(req, aliReq.Model) {
	case "video-edit":
		return applyHappyHorseVideoEdit(req, aliReq)
	case "r2v":
		return applyHappyHorseR2V(req, aliReq)
	case "i2v":
		return applyHappyHorseI2V(req, aliReq)
	default:
		return nil
	}
}

func applyHappyHorseI2V(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	urls := collectImageURLs(req, aliReq)
	if len(urls) == 0 {
		return fmt.Errorf("happyhorse i2v requires one first-frame image (images[], metadata.img_url, or content[].image_url)")
	}
	aliReq.Input.Media = []AliMediaItem{{Type: "first_frame", URL: urls[0]}}
	aliReq.Input.ImgURL = ""
	aliReq.Input.FirstFrameURL = ""
	// 官方图生视频不支持 ratio，宽高比跟随首帧
	if aliReq.Parameters != nil {
		aliReq.Parameters.Ratio = ""
	}
	return nil
}

func applyHappyHorseR2V(req relaycommon.TaskSubmitReq, aliReq *AliVideoRequest) error {
	urls := collectImageURLs(req, aliReq)
	if len(urls) == 0 {
		return fmt.Errorf("happyhorse r2v requires 1-9 reference images in images[] or metadata")
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
	aliReq.Parameters.Duration = 0
	if aliReq.Parameters.AudioSetting == "" {
		aliReq.Parameters.AudioSetting = "auto"
	}
	return nil
}

// mergeAliVideoMetadata 仅合并 metadata 中的 parameters / 扁平字段，避免整包 Unmarshal 覆盖 input.media。
func mergeAliVideoMetadata(metadata map[string]interface{}, aliReq *AliVideoRequest) error {
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
		if len(input.Media) > 0 {
			aliReq.Input.Media = input.Media
		}
		if input.ImgURL != "" {
			aliReq.Input.ImgURL = input.ImgURL
		}
		if input.FirstFrameURL != "" {
			aliReq.Input.FirstFrameURL = input.FirstFrameURL
		}
	}
	mergeFlatVideoMetadata(metadata, aliReq)
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
