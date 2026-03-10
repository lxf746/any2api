package zai_ocr

// OCRResponse is the top-level API response.
type OCRResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    OCRData  `json:"data"`
}

// OCRData holds the main result fields.
type OCRData struct {
	TaskID          string        `json:"task_id"`
	Status          string        `json:"status"`
	FileName        string        `json:"file_name"`
	FileSize        int64         `json:"file_size"`
	FileType        string        `json:"file_type"`
	FileURL         string        `json:"file_url"`
	CreatedAt       string        `json:"created_at"`
	MarkdownContent string        `json:"markdown_content"`
	JsonContentRaw  string        `json:"json_content"` // stringified JSON
	JsonContent     *JsonContent  `json:"-"`             // parsed from JsonContentRaw
	Layout          []LayoutBlock `json:"layout"`
	DataInfo        interface{}   `json:"data_info"`
	Timestamp       int64         `json:"timestamp"`
}

// JsonContent is the parsed structure inside json_content.
type JsonContent struct {
	MdResults     [][]MdResult     `json:"md_results"`
	LayoutDetails [][]LayoutDetail `json:"layout_details"`
	DataInfo      []PageInfo       `json:"data_info"`
	Usage         Usage            `json:"usage"`
}

// MdResult represents a single markdown block from OCR.
type MdResult struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// LayoutDetail represents a detected layout element with bounding box.
type LayoutDetail struct {
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	Position   []float64 `json:"position"`
	Confidence float64   `json:"confidence"`
}

// LayoutBlock is a top-level layout element in the response.
type LayoutBlock struct {
	Type       string      `json:"type"`
	SubType    string      `json:"sub_type"`
	Content    interface{} `json:"content"`
	BBox       []float64   `json:"bbox"`
	Order      int         `json:"order"`
	PageIdx    int         `json:"page_idx"`
}

// PageInfo holds page dimension info.
type PageInfo struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	PageNo int     `json:"page_no"`
}

// Usage holds token/page usage info.
type Usage struct {
	Pages  int `json:"pages"`
	Tokens int `json:"tokens"`
}

// AuthRequest is the request body for the auth endpoint.
type AuthRequest struct {
	Code string `json:"code"`
}

// AuthResponse is the response from POST /api/v1/z-ocr/auth/.
type AuthResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    AuthData `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

// AuthData holds the auth result fields.
type AuthData struct {
	UserID          string `json:"user_id"`
	AuthToken       string `json:"auth_token"`
	Name            string `json:"name"`
	ProfileImageURL string `json:"profile_image_url"`
}
