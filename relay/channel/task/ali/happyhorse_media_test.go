package ali

import (
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func TestCollectImageURLsFromImagesAndContent(t *testing.T) {
	req := relaycommon.TaskSubmitReq{
		Model:  "happyhorse-1.0-i2v",
		Images: []string{"data:image/jpeg;base64,abc"},
		Metadata: map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url": "https://example.com/a.png",
					},
				},
			},
		},
	}
	urls := collectImageURLs(req, &AliVideoRequest{})
	if len(urls) < 2 {
		t.Fatalf("expected >=2 urls, got %v", urls)
	}
}

func TestApplyHappyHorseI2VBuildsMedia(t *testing.T) {
	req := relaycommon.TaskSubmitReq{
		Model:  "happyhorse-1.0-i2v",
		Images: []string{"data:image/jpeg;base64,abc"},
	}
	aliReq := &AliVideoRequest{
		Model:      "happyhorse-1.0-i2v",
		Parameters: &AliVideoParameters{},
		Input:      AliVideoInput{Prompt: "run"},
	}
	if err := applyHappyHorseI2V(req, aliReq); err != nil {
		t.Fatal(err)
	}
	if len(aliReq.Input.Media) != 1 || aliReq.Input.Media[0].Type != "first_frame" {
		t.Fatalf("unexpected media: %+v", aliReq.Input.Media)
	}
	if aliReq.Input.Media[0].URL == "" {
		t.Fatal("empty media url")
	}
}

func TestIsHappyHorseTaskWithMappedUpstreamName(t *testing.T) {
	req := relaycommon.TaskSubmitReq{Model: "happyhorse-1.0-i2v"}
	if !isHappyHorseTask(req, "wan2.5-i2v-preview") {
		t.Fatal("expected happyhorse detection from req.Model")
	}
}
