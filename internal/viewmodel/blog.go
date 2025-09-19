package viewmodel

import (
	"cleanArch_with_postgres/internal/entity"
	"time"

	"gorm.io/gorm"
)

type BlogVM struct {
	Title      string         `json:"title"`
	Body       string         `json:"body"`
	Type       string         `json:"type"`
	Username   string         `json:"username"`
	Tags       string         `json:"tags"`
	Category   string         `json:"category"`
	Comments   []CommentVM    `json:"comments"`
	IsApproved bool           `json:"is_approved"`
	Status     string         `json:"status"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt"`
}

type BlogCreateVM struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Type     string `json:"type"`
	Tags     string `json:"tags"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

type BlogUpdateVM struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Type     string `json:"type"`
	Tags     string `json:"tags"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

type BlogUpdateResponse struct {
	Username  string    `json:"username"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Type      string    `json:"type"`
	Tags      string    `json:"tags"`
	Category  string    `json:"category"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToBlogVM(b *entity.Blog) *BlogVM {
	return &BlogVM{
		Title:      b.Content.Title,
		Body:       b.Content.Body,
		Type:       b.Content.Type,
		Username:   b.Content.Username,
		Tags:       b.Tags,
		Category:   b.Category,
		Comments:   ToCommentVMs(b.Comments),
		IsApproved: b.Content.IsApproved,
		Status:     b.Content.Status,
		CreatedAt:  b.CreatedAt,
		UpdatedAt:  b.UpdatedAt,
		DeletedAt:  b.DeletedAt,
	}
}

func ToBlogVMs(blogs []entity.Blog) []BlogVM {
	vms := make([]BlogVM, len(blogs))
	for i, b := range blogs {
		vms[i] = *ToBlogVM(&b)
	}
	return vms
}
