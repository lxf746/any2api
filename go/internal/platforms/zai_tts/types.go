package zai_tts

// TTSRequest is the request body for TTS generation.
type TTSRequest struct {
	VoiceName string  `json:"voice_name"`
	VoiceID   string  `json:"voice_id"`
	UserID    string  `json:"user_id"`
	InputText string  `json:"input_text"`
	Speed     float64 `json:"speed"`
	Volume    float64 `json:"volume"`
}

// AuthRequest is the request body for code → token exchange.
type AuthRequest struct {
	Code string `json:"code"`
}

// AuthResponse is the response from POST /api/v1/z-audio/auth/.
type AuthResponse struct {
	Code      int      `json:"code"`
	Message   string   `json:"message"`
	Data      AuthData `json:"data"`
	Timestamp int64    `json:"timestamp"`
}

// AuthData holds the auth result fields.
type AuthData struct {
	UserID          string `json:"user_id"`
	AuthToken       string `json:"auth_token"`
	Name            string `json:"name"`
	ProfileImageURL string `json:"profile_image_url"`
}
