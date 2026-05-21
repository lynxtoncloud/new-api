package ali

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func TestNormalizeHappyHorseDataURL(t *testing.T) {
	raw := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("png-bytes"))
	out, err := normalizeHappyHorseDataURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "data:image/png;base64,") {
		t.Fatalf("got %q", out)
	}
}

func TestEncodeHappyHorseHTTPImage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte{0xff, 0xd8, 0xff, 0x00})
	}))
	defer srv.Close()

	out, err := encodeHappyHorseImageToDataURL(srv.URL + "/a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "data:image/jpeg;base64,") {
		t.Fatalf("got %q", out)
	}
}

func TestFinalizeHappyHorseR2VEncodesMediaToDataURL(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte("jpeg-bytes"))
	req := relaycommon.TaskSubmitReq{
		Model:  "happyhorse-1.0-r2v",
		Prompt: "[Image 1] test",
		Images: []string{"data:image/jpeg;base64," + payload},
	}
	aliReq := &AliVideoRequest{
		Model:      "happyhorse-1.0-r2v",
		Parameters: &AliVideoParameters{Resolution: "720P", Ratio: "16:9", Duration: 5},
		Input:      AliVideoInput{Prompt: "[Image 1] test"},
	}
	if err := finalizeHappyHorseAliRequest(req, aliReq); err != nil {
		t.Fatal(err)
	}
	if len(aliReq.Input.Media) != 1 {
		t.Fatalf("media len=%d", len(aliReq.Input.Media))
	}
	if !strings.HasPrefix(aliReq.Input.Media[0].URL, "data:image/jpeg;base64,") {
		t.Fatalf("url=%q", aliReq.Input.Media[0].URL)
	}
	if aliReq.Parameters.Ratio != "16:9" {
		t.Fatalf("ratio=%q", aliReq.Parameters.Ratio)
	}
}
