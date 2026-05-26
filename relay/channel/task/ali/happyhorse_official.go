package ali

import "strings"

// 与百炼官方 HappyHorse 文档对齐的 parameters 取值（2026-05 文档）。
var (
	happyHorseOfficialResolutions = map[string]bool{
		"720P":  true,
		"1080P": true,
	}
	happyHorseT2VOfficialRatios = map[string]bool{
		"16:9": true, "9:16": true, "1:1": true,
		"4:3": true, "3:4": true, "4:5": true, "5:4": true,
	}
	happyHorseR2VOfficialRatios = map[string]bool{
		"16:9": true, "9:16": true, "3:4": true, "4:3": true,
		"4:5": true, "5:4": true, "1:1": true, "9:21": true, "21:9": true,
	}
)

func sanitizeHappyHorseRatioString(ratio string) string {
	r := strings.TrimSpace(ratio)
	r = strings.ReplaceAll(r, "：", ":")
	r = strings.ReplaceAll(r, "／", "/")
	r = strings.ReplaceAll(r, "/", ":")
	r = strings.ReplaceAll(r, " ", "")
	return r
}

func normalizeOfficialHappyHorseRatio(variant, ratio string) string {
	r := sanitizeHappyHorseRatioString(ratio)
	if r == "" {
		return "16:9"
	}
	switch variant {
	case "r2v":
		if happyHorseR2VOfficialRatios[r] {
			return r
		}
	case "t2v":
		if happyHorseT2VOfficialRatios[r] {
			return r
		}
	}
	return "16:9"
}

func normalizeOfficialHappyHorseResolution(res string) string {
	r := normalizeResolutionP(res)
	if happyHorseOfficialResolutions[r] {
		return r
	}
	return "1080P"
}

// normalizeOfficialHappyHorseParameters 按模式裁剪/规范化 parameters，避免万相遗留字段触发 InvalidParameter。
func normalizeOfficialHappyHorseParameters(aliReq *AliVideoRequest, variant string) {
	if aliReq == nil || aliReq.Parameters == nil {
		return
	}
	p := aliReq.Parameters
	p.Size = ""
	p.PromptExtend = false
	switch variant {
	case "t2v":
		if p.Resolution != "" {
			p.Resolution = normalizeOfficialHappyHorseResolution(p.Resolution)
		}
		p.Ratio = normalizeOfficialHappyHorseRatio("t2v", p.Ratio)
		p.AudioSetting = ""
	case "i2v":
		p.Ratio = ""
		p.Resolution = normalizeOfficialHappyHorseResolution(p.Resolution)
		if p.Resolution == "" {
			p.Resolution = "1080P"
		}
		if p.Duration <= 0 {
			p.Duration = 5
		}
		p.Duration = clampHappyHorseDuration(p.Duration)
		p.AudioSetting = ""
	case "r2v":
		p.Resolution = normalizeOfficialHappyHorseResolution(p.Resolution)
		if p.Resolution == "" {
			p.Resolution = "1080P"
		}
		p.Ratio = normalizeOfficialHappyHorseRatio("r2v", p.Ratio)
		if p.Duration <= 0 {
			p.Duration = 5
		}
		p.Duration = clampHappyHorseDuration(p.Duration)
		p.AudioSetting = ""
	case "video-edit":
		p.Ratio = ""
		p.Resolution = ""
		p.Duration = 0
		if p.AudioSetting == "" {
			p.AudioSetting = "auto"
		}
	default:
		p.Ratio = ""
	}
}
