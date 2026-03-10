package zai_image

// ImageRequest is the request body for image generation.
type ImageRequest struct {
	Prompt           string `json:"prompt"`
	Ratio            string `json:"ratio"`             // "9:16", "16:9", "1:1", "4:3", etc.
	Resolution       string `json:"resolution"`        // "1K", "2K", "4K"
	RmLabelWatermark bool   `json:"rm_label_watermark"`
}

// ImageResponse is the top-level API response.
type ImageResponse struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Data      ImageData `json:"data"`
	Timestamp int64     `json:"timestamp"`
}

// ImageData wraps the image result.
type ImageData struct {
	Image ImageInfo `json:"image"`
}

// ImageInfo holds the generated image details.
type ImageInfo struct {
	ImageID    string `json:"image_id"`
	Prompt     string `json:"prompt"`
	Size       string `json:"size"`       // "960x1728"
	Ratio      string `json:"ratio"`
	Resolution string `json:"resolution"`
	ImageURL   string `json:"image_url"`  // OSS signed URL, 7-day expiry
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

// AuthRequest is the request body for step 1: code → token.
type AuthRequest struct {
	Code string `json:"code"`
}

// AuthResponse is the response from POST /api/v1/z-image/auth/.
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

// CallbackRequest is the request body for step 2: register token as session.
type CallbackRequest struct {
	Token string `json:"token"`
}
