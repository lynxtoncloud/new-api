package doubao

import (
	"strconv"
	"strings"
)

var ModelList = []string{
	"doubao-seedance-1-0-pro-250528",
	"doubao-seedance-1-0-lite-t2v",
	"doubao-seedance-1-0-lite-i2v",
	"doubao-seedance-1-5-pro-251215",
	"doubao-seedance-2-0-260128",
	"doubao-seedance-2-0-fast-260128",
}

var ChannelName = "doubao-video"

// videoInputRatioMap 视频输入折扣比率（含视频单价 / 不含视频单价）。
// 管理员应将 ModelRatio 设置为"不含视频"的较高费率，
// 系统在检测到视频输入时自动乘以此折扣。
var videoInputRatioMap = map[string]float64{
	"doubao-seedance-2-0-260128":      28.0 / 46.0, // ~0.6087
	"doubao-seedance-2-0-fast-260128": 22.0 / 37.0, // ~0.5946
}

func GetVideoInputRatio(modelName string) (float64, bool) {
	r, ok := videoInputRatioMap[modelName]
	return r, ok
}

// resolutionRatioMap 分辨率倍率映射表，key 为分辨率字符串（大写 P），value 为倍率。
// 倍率 = 该分辨率单价 ÷ 基础分辨率（480P）单价。
// 管理员可按需调整具体数值；占位值均为 1.0，待你填入实际倍率。
var resolutionRatioMap = map[string]map[string]float64{
	"doubao-seedance-2-0-260128": {
		"480P":  1.0,         // TODO: 填入实际倍率
		"720P":  1.0,         // TODO: 填入实际倍率
		"1080P": 51.0 / 46.0, // TODO: 填入实际倍率
		"1440P": 1.0,         // TODO: 填入实际倍率
		"2160P": 1.0,         // TODO: 填入实际倍率
	},
	"doubao-seedance-2-0-fast-260128": {
		"480P":  1.0, // TODO: 填入实际倍率
		"720P":  1.0, // TODO: 填入实际倍率
		"1080P": 1.0, // TODO: 填入实际倍率
		"1440P": 1.0, // TODO: 填入实际倍率
		"2160P": 1.0, // TODO: 填入实际倍率
	},
}

func GetResolutionRatio(modelName, resolution string) (float64, bool) {
	modelRatios, ok := resolutionRatioMap[modelName]
	if !ok {
		return 0, false
	}
	// 统一为大写 P 格式（如 "720p" → "720P"）
	norm := strings.ToUpper(resolution)
	if !strings.HasSuffix(norm, "P") {
		// 尝试从像素尺寸转换（如 "1280x720" → "720P"）
		if r := pixelToResolution(norm); r != "" {
			norm = r
		}
	}
	ratio, ok := modelRatios[norm]
	return ratio, ok
}

// pixelToResolution 尝试将像素尺寸字符串转换为分辨率标签（如 "1280x720" → "720P"）。
func pixelToResolution(size string) string {
	parts := strings.SplitN(size, "X", 2)
	if len(parts) != 2 {
		return ""
	}
	w, errW := strconv.Atoi(parts[0])
	h, errH := strconv.Atoi(parts[1])
	if errW != nil || errH != nil {
		return ""
	}
	maxDim := w
	if h > maxDim {
		maxDim = h
	}
	switch {
	case maxDim >= 3840:
		return "2160P"
	case maxDim >= 1920:
		return "1080P"
	case maxDim >= 1280:
		return "720P"
	default:
		return "480P"
	}
}
