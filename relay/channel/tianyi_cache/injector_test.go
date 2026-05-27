package tianyi_cache

import (
	"bytes"
	"encoding/json"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func TestInjectExtraParamsOnNilParamOverride(t *testing.T) {
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{}}
	injectExtraParams(info)

	if info.ParamOverride == nil {
		t.Fatal("ParamOverride should be initialized")
	}
	if got, _ := info.ParamOverride[paramKeyUser].(string); got != paramValueUser {
		t.Fatalf("user mismatch: got %q want %q", got, paramValueUser)
	}
	raw, ok := info.ParamOverride[paramKeyCaching].(json.RawMessage)
	if !ok {
		t.Fatalf("caching should be json.RawMessage, got %T", info.ParamOverride[paramKeyCaching])
	}
	if !bytes.Equal(raw, paramValueCaching) {
		t.Fatalf("caching mismatch: got %s want %s", string(raw), string(paramValueCaching))
	}
}

func TestInjectExtraParamsOverridesExistingKeys(t *testing.T) {
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{}}
	info.ParamOverride = map[string]interface{}{
		paramKeyUser:    "client-supplied-user",
		paramKeyCaching: json.RawMessage(`{"type":"disabled"}`),
		"other":         "kept",
	}
	injectExtraParams(info)

	if got, _ := info.ParamOverride[paramKeyUser].(string); got != paramValueUser {
		t.Fatalf("user should be force-overridden: got %q", got)
	}
	raw, _ := info.ParamOverride[paramKeyCaching].(json.RawMessage)
	if !bytes.Equal(raw, paramValueCaching) {
		t.Fatalf("caching should be force-overridden: got %s", string(raw))
	}
	if got, _ := info.ParamOverride["other"].(string); got != "kept" {
		t.Fatalf("unrelated keys must be preserved, got %q", got)
	}
}

func TestInjectExtraParamsNilInfo(t *testing.T) {
	// 不应 panic
	injectExtraParams(nil)
}
