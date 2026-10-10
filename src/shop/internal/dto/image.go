package dto

type ImageRequest struct {
	Image []byte `json:"image" example:"iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"`
}

type ImageResponse struct {
	Image []byte `json:"image"`
}
