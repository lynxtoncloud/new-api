package ali

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "golang.org/x/image/webp"
)

const (
	happyHorseR2VMinShortSide   = 400 // 参考生视频：短边 ≥400px
	happyHorseI2VMinSide        = 300 // 图生视频首帧：宽、高均 ≥300px
	happyHorseMinAspectRatio    = 0.4 // 1:2.5
	happyHorseMaxAspectRatio    = 2.5
)

func validateHappyHorseImageBytes(data []byte, mediaType, label string) error {
	if len(data) == 0 {
		return fmt.Errorf("%s: empty image data", label)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%s: invalid image (%w)", label, err)
	}
	switch format {
	case "jpeg", "png", "webp":
	default:
		return fmt.Errorf("%s: unsupported format %q (use JPEG, PNG, or WEBP)", label, format)
	}
	typ := strings.ToLower(strings.TrimSpace(mediaType))
	switch typ {
	case "first_frame":
		if cfg.Width < happyHorseI2VMinSide || cfg.Height < happyHorseI2VMinSide {
			return fmt.Errorf("%s: first frame is %dx%dpx (DashScope i2v requires width and height at least %dpx)", label, cfg.Width, cfg.Height, happyHorseI2VMinSide)
		}
	default:
		short := cfg.Width
		long := cfg.Height
		if cfg.Height < short {
			short, long = cfg.Height, cfg.Width
		}
		if short < happyHorseR2VMinShortSide {
			return fmt.Errorf("%s: image short side is %dpx (DashScope r2v requires at least %dpx)", label, short, happyHorseR2VMinShortSide)
		}
		if long == 0 {
			return fmt.Errorf("%s: invalid image dimensions", label)
		}
	}
	short := cfg.Width
	long := cfg.Height
	if cfg.Height < short {
		short, long = cfg.Height, cfg.Width
	}
	if long == 0 {
		return fmt.Errorf("%s: invalid image dimensions", label)
	}
	ar := float64(long) / float64(short)
	if ar < happyHorseMinAspectRatio || ar > happyHorseMaxAspectRatio {
		return fmt.Errorf("%s: aspect ratio %.2f:1 is outside allowed range (1:2.5 to 2.5:1)", label, ar)
	}
	return nil
}

func validateHappyHorseImageDataURL(dataURL, mediaType, label string) error {
	dataURL = strings.TrimSpace(dataURL)
	comma := strings.Index(dataURL, ",")
	if comma < 0 {
		return fmt.Errorf("%s: invalid data url", label)
	}
	payload := stripBase64Whitespace(dataURL[comma+1:])
	raw, err := base64DecodeFlexible(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	return validateHappyHorseImageBytes(raw, mediaType, label)
}

func base64DecodeFlexible(payload string) ([]byte, error) {
	if b, err := base64StdDecode(payload); err == nil {
		return b, nil
	}
	return base64RawStdDecode(payload)
}
