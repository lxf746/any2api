package http

import (
	"encoding/json"
	"fmt"
	"math"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"any2api-go/internal/core"
	"any2api-go/internal/platforms"
	"any2api-go/internal/platforms/zai_image"
	"any2api-go/internal/platforms/zai_ocr"
	"any2api-go/internal/platforms/zai_tts"
)

var zaiImageSizeMap = map[string][2]string{
	"1024x1024": {"1:1", "1K"},
	"1024x1792": {"9:16", "2K"},
	"1792x1024": {"16:9", "2K"},
}

type Handler struct {
	registry *core.Registry
	cfg      core.AppConfig
	runtime  *core.RuntimeManager
	sessions *adminSessionStore
}

func NewHandler(registry *core.Registry, cfg core.AppConfig) *Handler {
	return &Handler{registry: registry, cfg: cfg}
}

func NewHandlerWithRuntime(runtime *core.RuntimeManager) *Handler {
	return &Handler{runtime: runtime, sessions: newAdminSessionStore()}
}

func (h *Handler) Routes() stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/v1/models", h.requireAPIKey(h.models))
	mux.HandleFunc("/v1/chat/completions", h.requireAPIKey(h.openAIChat))
	mux.HandleFunc("/v1/messages", h.requireAPIKey(h.anthropicMessages))
	mux.HandleFunc("/v1/images/generations", h.requireAPIKey(h.openAIImagesGeneration))
	mux.HandleFunc("/v1/audio/speech", h.requireAPIKey(h.openAIAudioSpeech))
	mux.HandleFunc("/v1/ocr", h.requireAPIKey(h.ocrUpload))
	if h.runtime != nil {
		h.registerAdminRoutes(mux)
	}
	return mux
}

func (h *Handler) health(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	h.writeJSON(w, 200, map[string]interface{}{"status": "ok", "project": "any2api-go"})
}

func (h *Handler) models(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	models, err := h.currentRegistry().Models(r.URL.Query().Get("provider"))
	if err != nil {
		h.writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	data := make([]map[string]interface{}, 0, len(models))
	for _, model := range models {
		data = append(data, map[string]interface{}{"id": model.PublicModel, "object": "model", "owned_by": model.OwnedBy, "provider": model.Provider, "upstream_model": model.UpstreamModel})
	}
	h.writeJSON(w, 200, map[string]interface{}{"object": "list", "data": data})
}

func (h *Handler) openAIChat(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req core.OpenAIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	provider, err := h.currentRegistry().Resolve(req.Provider)
	if err != nil {
		h.writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if !provider.Capabilities().OpenAICompatible {
		h.writeJSON(w, 400, map[string]string{"error": fmt.Sprintf("provider %q does not support OpenAI compatible endpoint", provider.ID())})
		return
	}
	w.Header().Set("X-Newplatform2API-Provider", provider.ID())
	unified := core.UnifiedRequest{ProviderHint: req.Provider, Protocol: "openai", Model: req.Model, Messages: req.Messages, Stream: req.Stream, ProviderOptions: h.providerOptionsFromHeaders(r)}
	if req.Stream {
		if streamProvider, ok := provider.(core.OpenAIStreamProvider); ok {
			events, err := streamProvider.StreamOpenAI(r.Context(), unified)
			if err != nil {
				h.writeJSON(w, 502, map[string]string{"error": err.Error()})
				return
			}
			h.writeOpenAIEventStream(w, req.Model, events)
			return
		}
	}
	reply := provider.GenerateReply(unified)
	if upstreamProvider, ok := provider.(core.OpenAIChatProvider); ok {
		text, err := upstreamProvider.CompleteOpenAI(r.Context(), unified)
		if err != nil {
			h.writeJSON(w, 502, map[string]string{"error": err.Error()})
			return
		}
		reply = text
	}
	if req.Stream {
		h.writeOpenAIStream(w, reply)
		return
	}
	h.writeJSON(w, 200, map[string]interface{}{"id": fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()), "object": "chat.completion", "created": time.Now().Unix(), "model": req.Model, "choices": []map[string]interface{}{{"index": 0, "message": map[string]interface{}{"role": "assistant", "content": reply}, "finish_reason": "stop"}}})
}

func (h *Handler) anthropicMessages(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req core.AnthropicMessagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	provider, err := h.currentRegistry().Resolve(req.Provider)
	if err != nil {
		h.writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if !provider.Capabilities().AnthropicCompatible {
		h.writeJSON(w, 400, map[string]string{"error": fmt.Sprintf("provider %q does not support Anthropic compatible endpoint", provider.ID())})
		return
	}
	w.Header().Set("X-Newplatform2API-Provider", provider.ID())
	unified := core.UnifiedRequest{ProviderHint: req.Provider, Protocol: "anthropic", Model: req.Model, Messages: req.Messages, System: req.System, Stream: req.Stream, ProviderOptions: h.providerOptionsFromHeaders(r)}
	if req.Stream {
		if streamProvider, ok := provider.(core.AnthropicStreamProvider); ok {
			events, err := streamProvider.StreamAnthropic(r.Context(), unified)
			if err != nil {
				h.writeJSON(w, 502, map[string]string{"error": err.Error()})
				return
			}
			h.writeAnthropicEventStream(w, req.Model, events)
			return
		}
	}
	reply := provider.GenerateReply(unified)
	if upstreamProvider, ok := provider.(core.AnthropicMessagesProvider); ok {
		text, err := upstreamProvider.CompleteAnthropic(r.Context(), unified)
		if err != nil {
			h.writeJSON(w, 502, map[string]string{"error": err.Error()})
			return
		}
		reply = text
	}
	if req.Stream {
		h.writeAnthropicStream(w, req.Model, reply)
		return
	}
	h.writeJSON(w, 200, map[string]interface{}{"id": fmt.Sprintf("msg_%d", time.Now().UnixNano()), "type": "message", "role": "assistant", "model": req.Model, "content": []map[string]interface{}{{"type": "text", "text": reply}}, "stop_reason": "end_turn"})
}

func (h *Handler) openAIImagesGeneration(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("X-Newplatform2API-Provider", "zai_image")
	payload, ok := h.readJSONObject(w, r)
	if !ok {
		return
	}
	prompt := strings.TrimSpace(stringValue(payload["prompt"]))
	if prompt == "" {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "prompt is required"})
		return
	}
	n, err := intValue(payload["n"], 1)
	if err != nil {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "n must be an integer"})
		return
	}
	if n != 1 {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "only n=1 is supported"})
		return
	}
	responseFormat := strings.ToLower(defaultString(payload["response_format"], "url"))
	if responseFormat != "" && responseFormat != "url" {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "only response_format=url is supported"})
		return
	}
	options, ok := h.providerOptionsFromPayload(w, payload)
	if !ok {
		return
	}
	ratio, resolution, ok := h.resolveImageSettings(w, payload, options)
	if !ok {
		return
	}
	rmLabelWatermark := coerceBool(firstNonNil(payload["rm_label_watermark"], options["rm_label_watermark"]), true)
	client := h.currentZAIImageClient(w)
	if client == nil {
		return
	}
	result, err := client.Generate(zai_image.ImageRequest{
		Prompt:           prompt,
		Ratio:            ratio,
		Resolution:       resolution,
		RmLabelWatermark: rmLabelWatermark,
	})
	if err != nil {
		h.writeJSON(w, stdhttp.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	created := result.Timestamp
	if created == 0 {
		created = time.Now().Unix()
	}
	image := result.Data.Image
	size := strings.TrimSpace(image.Size)
	if size == "" && image.Width > 0 && image.Height > 0 {
		size = fmt.Sprintf("%dx%d", image.Width, image.Height)
	}
	h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{
		"created": created,
		"data": []map[string]interface{}{{
			"url":            image.ImageURL,
			"revised_prompt": firstNonEmptyString(image.Prompt, prompt),
			"size":           size,
			"width":          image.Width,
			"height":         image.Height,
			"ratio":          firstNonEmptyString(image.Ratio, ratio),
			"resolution":     firstNonEmptyString(image.Resolution, resolution),
		}},
	})
}

func (h *Handler) openAIAudioSpeech(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("X-Newplatform2API-Provider", "zai_tts")
	payload, ok := h.readJSONObject(w, r)
	if !ok {
		return
	}
	text := firstNonEmptyString(stringValue(payload["input"]), stringValue(payload["text"]))
	if text == "" {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "input is required"})
		return
	}
	responseFormat := strings.ToLower(defaultString(payload["response_format"], "wav"))
	if responseFormat != "wav" {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "only response_format=wav is supported"})
		return
	}
	options, ok := h.providerOptionsFromPayload(w, payload)
	if !ok {
		return
	}
	voiceID := defaultString(payload["voice_id"], payload["voice"], options["voice_id"], "system_003")
	voiceName := defaultString(payload["voice_name"], options["voice_name"], "通用男声")
	speed, err := floatValue(firstNonNil(payload["speed"], options["speed"]), 1)
	if err != nil {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "speed and volume must be numbers"})
		return
	}
	volume, err := floatValue(firstNonNil(payload["volume"], options["volume"]), 1)
	if err != nil {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "speed and volume must be numbers"})
		return
	}
	client := h.currentZAITTSClient(w)
	if client == nil {
		return
	}
	audio, err := client.Synthesize(zai_tts.TTSRequest{
		VoiceName: voiceName,
		VoiceID:   voiceID,
		InputText: text,
		Speed:     speed,
		Volume:    volume,
	})
	if err != nil {
		h.writeJSON(w, stdhttp.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	h.writeBytes(w, stdhttp.StatusOK, audio, "audio/wav")
}

func (h *Handler) ocrUpload(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	w.Header().Set("X-Newplatform2API-Provider", "zai_ocr")
	if !strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data") {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "content-type must be multipart/form-data"})
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid multipart form-data"})
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()
	filename := "upload.bin"
	if fileHeader != nil && strings.TrimSpace(fileHeader.Filename) != "" {
		filename = strings.TrimSpace(fileHeader.Filename)
	}
	client := h.currentZAIOCRClient(w)
	if client == nil {
		return
	}
	result, err := client.ProcessReader(file, filename)
	if err != nil {
		h.writeJSON(w, stdhttp.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	jsonValue := structuredOCRJSON(result.Data.JsonContentRaw, result.Data.JsonContent)
	h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{
		"id":       result.Data.TaskID,
		"object":   "ocr.result",
		"model":    "zai-ocr",
		"status":   result.Data.Status,
		"text":     result.Data.MarkdownContent,
		"markdown": result.Data.MarkdownContent,
		"json":     jsonValue,
		"layout":   result.Data.Layout,
		"file": map[string]interface{}{
			"name":       result.Data.FileName,
			"size":       result.Data.FileSize,
			"type":       result.Data.FileType,
			"url":        result.Data.FileURL,
			"created_at": result.Data.CreatedAt,
		},
	})
}

func (h *Handler) writeJSON(w stdhttp.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (h *Handler) writeBytes(w stdhttp.ResponseWriter, status int, body []byte, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func (h *Handler) readJSONObject(w stdhttp.ResponseWriter, r *stdhttp.Request) (map[string]interface{}, bool) {
	var payload map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
		return nil, false
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	return payload, true
}

func (h *Handler) providerOptionsFromPayload(w stdhttp.ResponseWriter, payload map[string]interface{}) (map[string]interface{}, bool) {
	value, ok := payload["provider_options"]
	if !ok || value == nil {
		return map[string]interface{}{}, true
	}
	options, ok := value.(map[string]interface{})
	if !ok {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "provider_options must be an object"})
		return nil, false
	}
	return options, true
}

func (h *Handler) resolveImageSettings(w stdhttp.ResponseWriter, payload map[string]interface{}, options map[string]interface{}) (string, string, bool) {
	ratio := defaultString(payload["ratio"], options["ratio"])
	resolution := defaultString(payload["resolution"], options["resolution"])
	if ratio != "" && resolution != "" {
		return ratio, resolution, true
	}
	size := defaultString(payload["size"], options["size"])
	if size != "" {
		mapped, ok := zaiImageSizeMap[size]
		if !ok {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": fmt.Sprintf("unsupported size: %s", size)})
			return "", "", false
		}
		return mapped[0], mapped[1], true
	}
	if ratio == "" {
		ratio = "1:1"
	}
	if resolution == "" {
		resolution = "1K"
	}
	return ratio, resolution, true
}

func (h *Handler) currentZAIImageClient(w stdhttp.ResponseWriter) *zai_image.ImageClient {
	cfg := h.currentConfig()
	if strings.TrimSpace(cfg.ZAIImage.SessionToken) == "" {
		h.writeJSON(w, stdhttp.StatusServiceUnavailable, map[string]string{"error": "zai image is not configured"})
		return nil
	}
	client := zai_image.NewImageClient(cfg.ZAIImage.SessionToken)
	client.Endpoint = cfg.ZAIImage.APIURL
	return client
}

func (h *Handler) currentZAITTSClient(w stdhttp.ResponseWriter) *zai_tts.TTSClient {
	cfg := h.currentConfig()
	if strings.TrimSpace(cfg.ZAITTS.Token) == "" || strings.TrimSpace(cfg.ZAITTS.UserID) == "" {
		h.writeJSON(w, stdhttp.StatusServiceUnavailable, map[string]string{"error": "zai tts is not configured"})
		return nil
	}
	client := zai_tts.NewTTSClient(cfg.ZAITTS.Token, cfg.ZAITTS.UserID)
	client.Endpoint = cfg.ZAITTS.APIURL
	return client
}

func (h *Handler) currentZAIOCRClient(w stdhttp.ResponseWriter) *zai_ocr.OCRClient {
	cfg := h.currentConfig()
	if strings.TrimSpace(cfg.ZAIOCR.Token) == "" {
		h.writeJSON(w, stdhttp.StatusServiceUnavailable, map[string]string{"error": "zai ocr is not configured"})
		return nil
	}
	client := zai_ocr.NewOCRClient(cfg.ZAIOCR.Token)
	client.Endpoint = cfg.ZAIOCR.APIURL
	return client
}

func (h *Handler) providerOptionsFromHeaders(r *stdhttp.Request) map[string]string {
	options := map[string]string{}
	if value := strings.TrimSpace(r.Header.Get("X-Orchids-Client-Cookie")); value != "" {
		options["orchids_client_cookie"] = value
	}
	if value := strings.TrimSpace(r.Header.Get("X-Orchids-Client-UAT")); value != "" {
		options["orchids_client_uat"] = value
	}
	if value := strings.TrimSpace(r.Header.Get("X-Orchids-Session-Id")); value != "" {
		options["orchids_session_id"] = value
	}
	if value := strings.TrimSpace(r.Header.Get("X-Orchids-Project-Id")); value != "" {
		options["orchids_project_id"] = value
	}
	if value := strings.TrimSpace(r.Header.Get("X-Orchids-User-Id")); value != "" {
		options["orchids_user_id"] = value
	}
	if value := strings.TrimSpace(r.Header.Get("X-Orchids-Email")); value != "" {
		options["orchids_email"] = value
	}
	if value := strings.TrimSpace(r.Header.Get("X-Orchids-Agent-Mode")); value != "" {
		options["orchids_agent_mode"] = value
	}
	if len(options) == 0 {
		return nil
	}
	return options
}

func (h *Handler) requireAPIKey(next stdhttp.HandlerFunc) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if strings.TrimSpace(h.currentConfig().APIKey) == "" {
			next(w, r)
			return
		}
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(authHeader, "Bearer ") {
			h.writeJSON(w, stdhttp.StatusUnauthorized, map[string]interface{}{
				"error": map[string]string{
					"message": "Missing or invalid authorization header",
					"type":    "authentication_error",
					"code":    "missing_auth",
				},
			})
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token != h.currentConfig().APIKey {
			h.writeJSON(w, stdhttp.StatusUnauthorized, map[string]interface{}{
				"error": map[string]string{
					"message": "Invalid API key",
					"type":    "authentication_error",
					"code":    "invalid_api_key",
				},
			})
			return
		}
		next(w, r)
	}
}

func (h *Handler) currentConfig() core.AppConfig {
	if h.runtime != nil {
		return h.runtime.CurrentAppConfig()
	}
	return h.cfg
}

func (h *Handler) currentRegistry() *core.Registry {
	if h.runtime != nil {
		return platforms.DefaultRegistry(h.runtime.CurrentAppConfig())
	}
	return h.registry
}

func (h *Handler) writeOpenAIStream(w stdhttp.ResponseWriter, text string) {
	h.writeEventStreamHeaders(w)
	_, _ = fmt.Fprintf(w, "data: {\"id\":\"chatcmpl-skeleton\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":%q},\"finish_reason\":null}]}\n\n", text)
	_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	h.flush(w)
}

func (h *Handler) writeOpenAIEventStream(w stdhttp.ResponseWriter, model string, events <-chan core.TextStreamEvent) {
	if model == "" {
		model = "auto"
	}
	streamID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	h.writeEventStreamHeaders(w)
	_, _ = fmt.Fprintf(w, "data: {\"id\":%q,\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":%q,\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\"},\"finish_reason\":null}]}\n\n", streamID, time.Now().Unix(), model)
	h.flush(w)
	for event := range events {
		if event.Err != nil {
			break
		}
		if event.Delta == "" {
			continue
		}
		_, _ = fmt.Fprintf(w, "data: {\"id\":%q,\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":%q,\"choices\":[{\"index\":0,\"delta\":{\"content\":%q},\"finish_reason\":null}]}\n\n", streamID, time.Now().Unix(), model, event.Delta)
		h.flush(w)
	}
	_, _ = fmt.Fprintf(w, "data: {\"id\":%q,\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":%q,\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n", streamID, time.Now().Unix(), model)
	_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	h.flush(w)
}

func (h *Handler) writeAnthropicStream(w stdhttp.ResponseWriter, model string, text string) {
	output := make(chan core.TextStreamEvent, 1)
	output <- core.TextStreamEvent{Delta: text}
	close(output)
	h.writeAnthropicEventStream(w, model, output)
}

func (h *Handler) writeAnthropicEventStream(w stdhttp.ResponseWriter, model string, events <-chan core.TextStreamEvent) {
	if model == "" {
		model = "auto"
	}
	messageID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	h.writeEventStreamHeaders(w)
	_, _ = fmt.Fprintf(w, "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":%q,\"type\":\"message\",\"role\":\"assistant\",\"model\":%q,\"content\":[]}}\n\n", messageID, model)
	_, _ = fmt.Fprint(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n")
	h.flush(w)
	for event := range events {
		if event.Err != nil {
			_, _ = fmt.Fprintf(w, "event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"api_error\",\"message\":%q}}\n\n", event.Err.Error())
			h.flush(w)
			return
		}
		if event.Delta == "" {
			continue
		}
		_, _ = fmt.Fprintf(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":%q}}\n\n", event.Delta)
		h.flush(w)
	}
	_, _ = fmt.Fprint(w, "event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n")
	_, _ = fmt.Fprint(w, "event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\",\"stop_sequence\":null}}\n\n")
	_, _ = fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	h.flush(w)
}

func (h *Handler) writeEventStreamHeaders(w stdhttp.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

func (h *Handler) flush(w stdhttp.ResponseWriter) {
	if flusher, ok := w.(stdhttp.Flusher); ok {
		flusher.Flush()
	}
}

func firstNonNil(values ...interface{}) interface{} {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func defaultString(values ...interface{}) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(stringValue(value))
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func stringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func intValue(value interface{}, defaultValue int) (int, error) {
	if value == nil {
		return defaultValue, nil
	}
	switch v := value.(type) {
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i), nil
		}
		f, err := v.Float64()
		if err != nil || math.Trunc(f) != f {
			return 0, fmt.Errorf("invalid integer")
		}
		return int(f), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, fmt.Errorf("invalid integer")
		}
		return int(v), nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return defaultValue, nil
		}
		parsed, err := strconv.Atoi(trimmed)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("invalid integer")
	}
}

func floatValue(value interface{}, defaultValue float64) (float64, error) {
	if value == nil || value == "" {
		return defaultValue, nil
	}
	switch v := value.(type) {
	case json.Number:
		return v.Float64()
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return defaultValue, nil
		}
		return strconv.ParseFloat(trimmed, 64)
	default:
		return 0, fmt.Errorf("invalid number")
	}
}

func coerceBool(value interface{}, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i != 0
		}
	}
	return true
}

func structuredOCRJSON(raw string, parsed *zai_ocr.JsonContent) interface{} {
	if parsed != nil {
		return parsed
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	var value interface{}
	if err := json.Unmarshal([]byte(trimmed), &value); err == nil {
		return value
	}
	return raw
}
