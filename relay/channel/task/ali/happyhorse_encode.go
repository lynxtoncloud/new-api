package ali

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/service"
)

const happyHorseMaxImageBytes = 20 << 20 // 官方：单图不超过 20MB

var happyHorseImageHTTPClient = &http.Client{Timeout: 45 * time.Second}

// encodeHappyHorseMediaImagesToBase64 将 input.media 中的图片统一为官方 data:{mime};base64,{data} 格式。
// video-edit 的 video 条目保持公网 URL，不做 Base64（体积过大且文档要求视频为 URL）。
func encodeHappyHorseMediaImagesToBase64(media []AliMediaItem) ([]AliMediaItem, error) {
	if len(media) == 0 {
		return media, nil
	}
	out := make([]AliMediaItem, len(media))
	for i, item := range media {
		typ := strings.TrimSpace(item.Type)
		rawURL := strings.TrimSpace(item.URL)
		switch typ {
		case "video":
			if !isPublicHTTPURL(rawURL) {
				return nil, fmt.Errorf("happyhorse %s: video url must be public http(s)", typ)
			}
			out[i] = AliMediaItem{Type: typ, URL: rawURL}
		case "first_frame", "reference_image":
			encoded, err := encodeHappyHorseImageToDataURL(rawURL)
			if err != nil {
				return nil, fmt.Errorf("happyhorse encode %s[%d]: %w", typ, i, err)
			}
			out[i] = AliMediaItem{Type: typ, URL: encoded}
		default:
			return nil, fmt.Errorf("happyhorse: unsupported media type %q", typ)
		}
	}
	return out, nil
}

func encodeHappyHorseImageToDataURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty image url")
	}
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "blob:") {
		return "", fmt.Errorf("blob urls are not supported; upload images via the API body")
	}
	if strings.HasPrefix(lower, "data:") {
		return normalizeHappyHorseDataURL(raw)
	}
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return downloadHappyHorseImageAsDataURL(raw)
	}
	if looksLikeRawBase64(raw) {
		return fmt.Sprintf("data:image/jpeg;base64,%s", stripBase64Whitespace(raw)), nil
	}
	return "", fmt.Errorf("unsupported image reference (need data: url, http(s) url, or base64)")
}

func normalizeHappyHorseDataURL(dataURL string) (string, error) {
	dataURL = strings.TrimSpace(dataURL)
	comma := strings.Index(dataURL, ",")
	if comma < 0 {
		return "", fmt.Errorf("invalid data url: missing payload")
	}
	header := strings.TrimSpace(dataURL[:comma])
	payload := stripBase64Whitespace(dataURL[comma+1:])
	if payload == "" {
		return "", fmt.Errorf("invalid data url: empty base64 payload")
	}
	if !strings.HasPrefix(strings.ToLower(header), "data:") {
		return "", fmt.Errorf("invalid data url header")
	}
	mimePart := strings.TrimPrefix(header, "data:")
	mimePart = strings.TrimPrefix(mimePart, "DATA:")
	mimePart = strings.TrimSuffix(mimePart, ";base64")
	mimePart = strings.TrimSuffix(mimePart, ";BASE64")
	mime := normalizeHappyHorseImageMIME(mimePart)
	if _, err := base64.StdEncoding.DecodeString(payload); err != nil {
		if _, err2 := base64.RawStdEncoding.DecodeString(payload); err2 != nil {
			return "", fmt.Errorf("invalid base64 in data url: %w", err)
		}
	}
	// 官方格式：data:{MIME};base64,{data}
	return fmt.Sprintf("data:%s;base64,%s", mime, payload), nil
}

func downloadHappyHorseImageAsDataURL(url string) (string, error) {
	client := happyHorseImageHTTPClient
	if c := service.GetHttpClient(); c != nil && c.Timeout > 0 {
		client = c
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download image http %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, happyHorseMaxImageBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", err
	}
	if len(body) > happyHorseMaxImageBytes {
		return "", fmt.Errorf("image exceeds 20MB limit")
	}
	if len(body) == 0 {
		return "", fmt.Errorf("empty image body")
	}
	ct := resp.Header.Get("Content-Type")
	mime := normalizeHappyHorseImageMIME(ct)
	if !strings.Contains(strings.ToLower(ct), "image/") {
		mime = guessMIMEFromURL(url, "image/jpeg")
	}
	encoded := base64.StdEncoding.EncodeToString(body)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded), nil
}

func normalizeHappyHorseImageMIME(mime string) string {
	mime = strings.ToLower(strings.TrimSpace(mime))
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = mime[:i]
	}
	switch mime {
	case "image/jpg", "image/jpeg":
		return "image/jpeg"
	case "image/png":
		return "image/png"
	case "image/webp":
		return "image/webp"
	case "":
		return "image/jpeg"
	default:
		if strings.HasPrefix(mime, "image/") {
			return mime
		}
		return "image/jpeg"
	}
}

func guessMIMEFromURL(url, fallback string) string {
	lower := strings.ToLower(url)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	default:
		return fallback
	}
}

func stripBase64Whitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func looksLikeRawBase64(s string) bool {
	s = stripBase64Whitespace(s)
	if len(s) < 32 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '+', r == '/', r == '=':
			continue
		default:
			return false
		}
	}
	return true
}
