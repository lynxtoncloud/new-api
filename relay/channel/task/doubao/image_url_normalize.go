package doubao

import (
	"encoding/base64"
	"strings"
)

// normalizeDoubaoImageURLString 将裸 base64（无前缀）包装为 data URL，便于火山方舟 image_url.url 识别。
// 已支持的格式原样返回：http(s)、data:、asset://、oss://。
func normalizeDoubaoImageURLString(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	low := strings.ToLower(s)
	if strings.HasPrefix(low, "http://") || strings.HasPrefix(low, "https://") {
		return s
	}
	if strings.HasPrefix(low, "data:") {
		return s
	}
	if strings.HasPrefix(low, "asset://") || strings.HasPrefix(low, "oss://") {
		return s
	}
	clean := stripDoubaoBase64Whitespace(s)
	if len(clean) < 32 {
		return s
	}
	decoded, err := base64.StdEncoding.DecodeString(clean)
	if err != nil || len(decoded) < 8 {
		return s
	}
	prefix := sniffDoubaoImageDataURLPrefix(decoded)
	return prefix + clean
}

func stripDoubaoBase64Whitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

func sniffDoubaoImageDataURLPrefix(b []byte) string {
	if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xD8 {
		return "data:image/jpeg;base64,"
	}
	if len(b) >= 8 && b[0] == 0x89 && b[1] == 0x50 && b[2] == 0x4E && b[3] == 0x47 && b[4] == 0x0D && b[5] == 0x0A && b[6] == 0x1A && b[7] == 0x0A {
		return "data:image/png;base64,"
	}
	if len(b) >= 6 {
		h := string(b[:6])
		if h == "GIF87a" || h == "GIF89a" {
			return "data:image/gif;base64,"
		}
	}
	if len(b) >= 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WEBP" {
		return "data:image/webp;base64,"
	}
	return "data:image/jpeg;base64,"
}

func normalizeDoubaoContentImageURLs(content *[]ContentItem) {
	if content == nil {
		return
	}
	items := *content
	for i := range items {
		if items[i].Type != "image_url" || items[i].ImageURL == nil {
			continue
		}
		items[i].ImageURL.URL = normalizeDoubaoImageURLString(items[i].ImageURL.URL)
	}
	*content = items
}
