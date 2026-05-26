package ali

import (
	"strings"
	"testing"
)

func TestValidateHappyHorseR2VRejectsSmallImage(t *testing.T) {
	dataURL := happyHorseTestJPEGDataURL(200, 200)
	err := validateHappyHorseImageDataURL(dataURL, "reference_image", "test")
	if err == nil {
		t.Fatal("expected error for short side < 400px")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateHappyHorseI2VFirstFrameMinSide(t *testing.T) {
	okURL := happyHorseTestJPEGDataURL(300, 500)
	if err := validateHappyHorseImageDataURL(okURL, "first_frame", "test"); err != nil {
		t.Fatalf("300x500 should pass i2v: %v", err)
	}
	smallURL := happyHorseTestJPEGDataURL(299, 400)
	if err := validateHappyHorseImageDataURL(smallURL, "first_frame", "test"); err == nil {
		t.Fatal("expected error when width < 300px")
	}
}

func TestEnsureHappyHorseR2VPromptAddsImageTags(t *testing.T) {
	got := ensureHappyHorseR2VPrompt("奔跑", 2)
	if !strings.Contains(got, "[Image 1]") || !strings.Contains(got, "[Image 2]") {
		t.Fatalf("got %q", got)
	}
	if !strings.Contains(got, "奔跑") {
		t.Fatalf("prompt lost: %q", got)
	}
}
