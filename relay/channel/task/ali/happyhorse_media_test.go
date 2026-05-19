package ali

import (
	"encoding/json"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func TestFinalizeHappyHorseI2VMarshalsMedia(t *testing.T) {
	req := relaycommon.TaskSubmitReq{
		Model:  "happyhorse-1.0-i2v",
		Prompt: "run",
		Images: []string{"data:image/jpeg;base64,abc"},
	}
	aliReq := &AliVideoRequest{
		Model:      "wan2.5-i2v-preview",
		Parameters: &AliVideoParameters{},
		Input:      AliVideoInput{Prompt: "run"},
	}
	if err := finalizeHappyHorseAliRequest(req, aliReq); err != nil {
		t.Fatal(err)
	}
	if aliReq.Model != "happyhorse-1.0-i2v" {
		t.Fatalf("model=%q want happyhorse-1.0-i2v", aliReq.Model)
	}
	raw, err := marshalAliVideoRequestBody(aliReq, req)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	input, ok := body["input"].(map[string]interface{})
	if !ok {
		t.Fatalf("input missing: %s", string(raw))
	}
	media, ok := input["media"].([]interface{})
	if !ok || len(media) == 0 {
		t.Fatalf("media missing in json: %s", string(raw))
	}
}

func TestFinalizeHappyHorseR2VMultipleImages(t *testing.T) {
	req := relaycommon.TaskSubmitReq{
		Model:  "happyhorse-1.0-r2v",
		Prompt: "ref",
		Images: []string{"https://a/1.png", "https://a/2.png"},
	}
	aliReq := &AliVideoRequest{
		Model:      "happyhorse-1.0-r2v",
		Parameters: &AliVideoParameters{Resolution: "1080P"},
		Input:      AliVideoInput{Prompt: "ref"},
	}
	if err := finalizeHappyHorseAliRequest(req, aliReq); err != nil {
		t.Fatal(err)
	}
	if len(aliReq.Input.Media) != 2 {
		t.Fatalf("media len=%d", len(aliReq.Input.Media))
	}
	if aliReq.Input.Media[0].Type != "reference_image" {
		t.Fatalf("type=%q", aliReq.Input.Media[0].Type)
	}
}

func TestFinalizeHappyHorseFailsWithoutImages(t *testing.T) {
	req := relaycommon.TaskSubmitReq{Model: "happyhorse-1.0-i2v", Prompt: "x"}
	aliReq := &AliVideoRequest{
		Model:      "happyhorse-1.0-i2v",
		Parameters: &AliVideoParameters{},
		Input:      AliVideoInput{Prompt: "x"},
	}
	if err := finalizeHappyHorseAliRequest(req, aliReq); err == nil {
		t.Fatal("expected error without images")
	}
}
