package entity

type Content struct {
	Title      string `json:"title"`
	Body       string `json:"body"`
	AuthorID   int    `json:"author_id"`
	Username   string `json:"username"`
	Type       string `json:"type"`
	IsApproved bool   `json:"is_approved"`
	Status     string `json:"status"`
}
