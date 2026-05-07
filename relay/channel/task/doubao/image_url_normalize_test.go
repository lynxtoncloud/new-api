package doubao

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestNormalizeDoubaoImageURLString(t *testing.T) {
	// 足够长的 JPEG 头 + 填充，使 base64 串长度 > 32
	jpegLike := make([]byte, 64)
	jpegLike[0] = 0xff
	jpegLike[1] = 0xd8
	for i := 2; i < len(jpegLike); i++ {
		jpegLike[i] = byte(i)
	}
	b64 := base64.StdEncoding.EncodeToString(jpegLike)
	out := normalizeDoubaoImageURLString(b64)
	if !strings.HasPrefix(out, "data:image/jpeg;base64,") {
		t.Fatalf("expected jpeg data URL prefix, got %.60q", out)
	}

	httpsURL := "https://example.com/a.jpg"
	if normalizeDoubaoImageURLString(httpsURL) != httpsURL {
		t.Fatal("https URL should pass through")
	}
	data := "data:image/jpeg;base64,abc"
	if normalizeDoubaoImageURLString(data) != data {
		t.Fatal("data: URL should pass through")
	}
}
