package ali

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
)

func happyHorseTestJPEGDataURL(width, height int) string {
	if width < 1 {
		width = 400
	}
	if height < 1 {
		height = 400
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}
