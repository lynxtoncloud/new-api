package ali

import (
	"encoding/json"
	"strings"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func TestFinalizeHappyHorseI2VParametersNoRatio(t *testing.T) {
	req := relaycommon.TaskSubmitReq{
		Model:  "happyhorse-1.0-i2v",
		Prompt: "猫在草地上奔跑",
		Images: []string{happyHorseTestJPEGDataURL(640, 480)},
		Metadata: map[string]interface{}{
			"ratio":        "16:9",
			"aspect_ratio": "16:9",
		},
	}
	aliReq := &AliVideoRequest{
		Model:      "happyhorse-1.0-i2v",
		Parameters: &AliVideoParameters{Ratio: "16:9", Resolution: "720P", Duration: 8},
		Input:      AliVideoInput{Prompt: "猫在草地上奔跑"},
	}
	if err := finalizeHappyHorseAliRequest(req, aliReq); err != nil {
		t.Fatal(err)
	}
	raw, err := marshalAliVideoRequestBody(aliReq, req)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"ratio"`) {
		t.Fatalf("i2v must not send ratio: %s", string(raw))
	}
	var body map[string]interface{}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	params, ok := body["parameters"].(map[string]interface{})
	if !ok {
		t.Fatalf("parameters missing: %s", string(raw))
	}
	if params["resolution"] != "720P" {
		t.Fatalf("resolution=%v", params["resolution"])
	}
	if params["duration"] != float64(8) && params["duration"] != 8 {
		t.Fatalf("duration=%v", params["duration"])
	}
	input, ok := body["input"].(map[string]interface{})
	if !ok {
		t.Fatal("input missing")
	}
	media, ok := input["media"].([]interface{})
	if !ok || len(media) != 1 {
		t.Fatalf("expected single first_frame media, got %v", input["media"])
	}
	m0, ok := media[0].(map[string]interface{})
	if !ok || m0["type"] != "first_frame" {
		t.Fatalf("media[0]=%v", media[0])
	}
}

func TestFinalizeHappyHorseI2VUsesOnlyFirstImage(t *testing.T) {
	req := relaycommon.TaskSubmitReq{
		Model:  "happyhorse-1.0-i2v",
		Prompt: "run",
		Images: []string{
			happyHorseTestJPEGDataURL(400, 400),
			happyHorseTestJPEGDataURL(500, 500),
		},
	}
	aliReq := &AliVideoRequest{
		Model:      "happyhorse-1.0-i2v",
		Parameters: &AliVideoParameters{},
		Input:      AliVideoInput{Prompt: "run"},
	}
	if err := finalizeHappyHorseAliRequest(req, aliReq); err != nil {
		t.Fatal(err)
	}
	if len(aliReq.Input.Media) != 1 {
		t.Fatalf("media len=%d want 1", len(aliReq.Input.Media))
	}
}
