package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

type HasPrompt interface {
	GetPrompt() string
}

type HasImage interface {
	HasImage() bool
}

func GetFullRequestURL(baseURL string, requestURL string, channelType int) string {
	fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestURL)

	if strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") {
		switch channelType {
		case constant.ChannelTypeOpenAI:
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(requestURL, "/v1"))
		case constant.ChannelTypeAzure:
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(requestURL, "/openai/deployments"))
		}
	}
	return fullRequestURL
}

func GetAPIVersion(c *gin.Context) string {
	query := c.Request.URL.Query()
	apiVersion := query.Get("api-version")
	if apiVersion == "" {
		apiVersion = c.GetString("api_version")
	}
	return apiVersion
}

func createTaskError(err error, code string, statusCode int, localError bool) *dto.TaskError {
	return &dto.TaskError{
		Code:       code,
		Message:    err.Error(),
		StatusCode: statusCode,
		LocalError: localError,
		Error:      err,
	}
}

func storeTaskRequest(c *gin.Context, info *RelayInfo, action string, requestObj TaskSubmitReq) {
	info.Action = action
	c.Set("task_request", requestObj)
}
func GetTaskRequest(c *gin.Context) (TaskSubmitReq, error) {
	v, exists := c.Get("task_request")
	if !exists {
		return TaskSubmitReq{}, fmt.Errorf("request not found in context")
	}
	req, ok := v.(TaskSubmitReq)
	if !ok {
		return TaskSubmitReq{}, fmt.Errorf("invalid task request type")
	}
	return req, nil
}

// LoadTaskSubmitReqForUpstream 合并校验阶段存入 context 的请求与 body 重载结果，避免二次解析丢图。
func LoadTaskSubmitReqForUpstream(c *gin.Context) (TaskSubmitReq, error) {
	var stored TaskSubmitReq
	hasStored := false
	if req, err := GetTaskRequest(c); err == nil {
		stored = req
		hasStored = true
	}
	reloaded, err := ReloadTaskSubmitReq(c)
	if err != nil {
		if hasStored {
			return stored, nil
		}
		return TaskSubmitReq{}, err
	}
	if !hasStored {
		return reloaded, nil
	}
	return MergeTaskSubmitReq(stored, reloaded), nil
}

// MergeTaskSubmitReq 合并两次解析的任务请求，优先保留已有字段并补全 images/metadata。
func MergeTaskSubmitReq(base, extra TaskSubmitReq) TaskSubmitReq {
	out := base
	if strings.TrimSpace(out.Model) == "" {
		out.Model = extra.Model
	}
	if strings.TrimSpace(out.Prompt) == "" {
		out.Prompt = extra.Prompt
	}
	if out.Duration == 0 {
		out.Duration = extra.Duration
	}
	if strings.TrimSpace(out.InputReference) == "" {
		out.InputReference = extra.InputReference
	}
	seen := map[string]bool{}
	var images []string
	addImg := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			return
		}
		seen[u] = true
		images = append(images, u)
	}
	for _, u := range out.Images {
		addImg(u)
	}
	for _, u := range extra.Images {
		addImg(u)
	}
	out.Images = images
	if out.Metadata == nil {
		out.Metadata = map[string]interface{}{}
	}
	for k, v := range extra.Metadata {
		if _, ok := out.Metadata[k]; !ok {
			out.Metadata[k] = v
		}
	}
	out.Normalize()
	return out
}

// ReloadTaskSubmitReq 从 body 完整重载任务请求（含 images/content），供转发上游前使用。
func ReloadTaskSubmitReq(c *gin.Context) (TaskSubmitReq, error) {
	var req TaskSubmitReq
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		return req, err
	}
	if req.InputReference != "" && len(req.Images) == 0 {
		req.Images = []string{req.InputReference}
	}
	req.Normalize()
	if strings.Contains(strings.ToLower(req.Model), "happyhorse") {
		if storage, err := common.GetBodyStorage(c); err == nil && storage != nil {
			if raw, err := storage.Bytes(); err == nil && len(raw) > 0 {
				mergeTaskImagesFromRawJSON(&req, raw)
			}
		}
		req.Normalize()
	}
	return req, nil
}

func validatePrompt(prompt string) *dto.TaskError {
	if strings.TrimSpace(prompt) == "" {
		return createTaskError(fmt.Errorf("prompt is required"), "invalid_request", http.StatusBadRequest, true)
	}
	return nil
}

func validateMultipartTaskRequest(c *gin.Context, info *RelayInfo, action string) (TaskSubmitReq, error) {
	var req TaskSubmitReq
	if _, err := c.MultipartForm(); err != nil {
		return req, err
	}

	formData := c.Request.PostForm
	req = TaskSubmitReq{
		Prompt:   formData.Get("prompt"),
		Model:    formData.Get("model"),
		Mode:     formData.Get("mode"),
		Image:    formData.Get("image"),
		Size:     formData.Get("size"),
		Metadata: make(map[string]interface{}),
	}

	if durationStr := formData.Get("seconds"); durationStr != "" {
		if duration, err := strconv.Atoi(durationStr); err == nil {
			req.Duration = duration
		}
	}

	if images := formData["images"]; len(images) > 0 {
		req.Images = images
	}

	for key, values := range formData {
		if len(values) > 0 && !isKnownTaskField(key) {
			if intVal, err := strconv.Atoi(values[0]); err == nil {
				req.Metadata[key] = intVal
			} else if floatVal, err := strconv.ParseFloat(values[0], 64); err == nil {
				req.Metadata[key] = floatVal
			} else {
				req.Metadata[key] = values[0]
			}
		}
	}
	return req, nil
}

func ValidateMultipartDirect(c *gin.Context, info *RelayInfo) *dto.TaskError {
	var prompt string
	var model string
	var seconds int
	var size string
	var hasInputReference bool

	var req TaskSubmitReq
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		return createTaskError(err, "invalid_json", http.StatusBadRequest, true)
	}

	prompt = req.Prompt
	model = req.Model
	size = req.Size
	seconds, _ = strconv.Atoi(req.Seconds)
	if seconds == 0 {
		seconds = req.Duration
	}
	if req.InputReference != "" {
		req.Images = []string{req.InputReference}
	}
	req.Normalize()

	if strings.TrimSpace(req.Model) == "" {
		return createTaskError(fmt.Errorf("model field is required"), "missing_model", http.StatusBadRequest, true)
	}

	// 欢乐马：从原始 body 再补全 images，并在入队前校验
	if strings.Contains(strings.ToLower(req.Model), "happyhorse") {
		if storage, err := common.GetBodyStorage(c); err == nil && storage != nil {
			if raw, err := storage.Bytes(); err == nil && len(raw) > 0 {
				mergeTaskImagesFromRawJSON(&req, raw)
			}
		}
		req.Normalize()
		if taskErr := validateHappyHorseTaskSubmitReq(req); taskErr != nil {
			return taskErr
		}
	}

	if req.HasImage() {
		hasInputReference = true
	}

	if taskErr := validatePrompt(prompt); taskErr != nil {
		return taskErr
	}

	action := constant.TaskActionTextGenerate
	if hasInputReference {
		action = constant.TaskActionGenerate
	}
	if strings.HasPrefix(model, "sora-2") {

		if size == "" {
			size = "720x1280"
		}

		if seconds <= 0 {
			seconds = 4
		}

		if model == "sora-2" && !lo.Contains([]string{"720x1280", "1280x720"}, size) {
			return createTaskError(fmt.Errorf("sora-2 size is invalid"), "invalid_size", http.StatusBadRequest, true)
		}
		if model == "sora-2-pro" && !lo.Contains([]string{"720x1280", "1280x720", "1792x1024", "1024x1792"}, size) {
			return createTaskError(fmt.Errorf("sora-2 size is invalid"), "invalid_size", http.StatusBadRequest, true)
		}
		// OtherRatios 已移到 Sora adaptor 的 EstimateBilling 中设置
	}

	storeTaskRequest(c, info, action, req)

	return nil
}

func appendImagesFromMetadataInput(req *TaskSubmitReq, meta map[string]interface{}) {
	if req == nil || meta == nil {
		return
	}
	inputRaw, ok := meta["input"].(map[string]interface{})
	if !ok {
		return
	}
	seen := map[string]bool{}
	for _, img := range req.Images {
		seen[img] = true
	}
	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			return
		}
		seen[u] = true
		req.Images = append(req.Images, u)
	}
	if mediaRaw, ok := inputRaw["media"].([]interface{}); ok {
		for _, item := range mediaRaw {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			typ, _ := m["type"].(string)
			typ = strings.ToLower(strings.TrimSpace(typ))
			if typ == "first_frame" || typ == "reference_image" || typ == "image" {
				if url, ok := m["url"].(string); ok {
					add(url)
				}
			}
		}
	}
	for _, key := range []string{"img_url", "first_frame_url", "image_url"} {
		if url, ok := inputRaw[key].(string); ok {
			add(url)
		}
	}
}

func mergeTaskImagesFromRawJSON(req *TaskSubmitReq, raw []byte) {
	if req == nil || len(raw) == 0 {
		return
	}
	var probe struct {
		Images         []string        `json:"images"`
		Image          string          `json:"image"`
		InputReference string          `json:"input_reference"`
		Metadata       json.RawMessage `json:"metadata"`
		Content        json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return
	}
	seen := map[string]bool{}
	for _, u := range req.Images {
		seen[u] = true
	}
	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			return
		}
		seen[u] = true
		req.Images = append(req.Images, u)
	}
	for _, u := range probe.Images {
		add(u)
	}
	add(probe.Image)
	add(probe.InputReference)
	if len(probe.Metadata) > 0 {
		var meta map[string]interface{}
		if json.Unmarshal(probe.Metadata, &meta) == nil {
			if req.Metadata == nil {
				req.Metadata = meta
			} else {
				for k, v := range meta {
					if _, ok := req.Metadata[k]; !ok {
						req.Metadata[k] = v
					}
				}
			}
			appendImagesFromMetadataInput(req, meta)
		}
	}
	if len(probe.Content) > 0 {
		var content []interface{}
		if json.Unmarshal(probe.Content, &content) == nil {
			if req.Metadata == nil {
				req.Metadata = map[string]interface{}{}
			}
			req.Metadata["content"] = content
		}
	}
}

func validateHappyHorseTaskSubmitReq(req TaskSubmitReq) *dto.TaskError {
	m := strings.ToLower(strings.TrimSpace(req.Model))
	if !strings.Contains(m, "happyhorse") {
		return nil
	}
	switch {
	case strings.Contains(m, "video-edit"):
		if collectHappyHorseVideoURL(req) == "" {
			return createTaskError(fmt.Errorf("happyhorse video-edit requires reference_video_url"), "missing_video", http.StatusBadRequest, true)
		}
	case strings.Contains(m, "i2v"), strings.Contains(m, "r2v"):
		if len(req.Images) == 0 {
			return createTaskError(fmt.Errorf("happyhorse %s requires at least one image in images[]", m), "missing_images", http.StatusBadRequest, true)
		}
	}
	return nil
}

func collectHappyHorseVideoURL(req TaskSubmitReq) string {
	if u := strings.TrimSpace(req.ReferenceVideoURL); u != "" {
		return u
	}
	if len(req.ReferenceVideoURLs) > 0 {
		if u := strings.TrimSpace(req.ReferenceVideoURLs[0]); u != "" {
			return u
		}
	}
	if req.Metadata != nil {
		if v, ok := req.Metadata["reference_video_url"].(string); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func isKnownTaskField(field string) bool {
	knownFields := map[string]bool{
		"prompt":          true,
		"model":           true,
		"mode":            true,
		"image":           true,
		"images":          true,
		"size":            true,
		"duration":        true,
		"input_reference": true, // Sora 特有字段
	}
	return knownFields[field]
}

func ValidateBasicTaskRequest(c *gin.Context, info *RelayInfo, action string) *dto.TaskError {
	var err error
	contentType := c.GetHeader("Content-Type")
	var req TaskSubmitReq
	if strings.HasPrefix(contentType, "multipart/form-data") {
		req, err = validateMultipartTaskRequest(c, info, action)
		if err != nil {
			return createTaskError(err, "invalid_multipart_form", http.StatusBadRequest, true)
		}
	}
	// 为了metadata字段的兼容性，统一UnmarshalBodyReusable
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		return createTaskError(err, "invalid_request", http.StatusBadRequest, true)
	}

	if taskErr := validatePrompt(req.Prompt); taskErr != nil {
		return taskErr
	}

	if len(req.Images) == 0 && strings.TrimSpace(req.Image) != "" {
		// 兼容单图上传
		req.Images = []string{req.Image}
	}
	req.Normalize()
	if strings.Contains(strings.ToLower(req.Model), "happyhorse") {
		if storage, err := common.GetBodyStorage(c); err == nil && storage != nil {
			if raw, err := storage.Bytes(); err == nil && len(raw) > 0 {
				mergeTaskImagesFromRawJSON(&req, raw)
			}
		}
		req.Normalize()
		if taskErr := validateHappyHorseTaskSubmitReq(req); taskErr != nil {
			return taskErr
		}
	}

	storeTaskRequest(c, info, action, req)
	return nil
}
