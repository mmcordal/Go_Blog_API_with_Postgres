package viewmodel

import (
	"cleanArch_with_postgres/internal/entity"
	"time"
)

type CommentVM struct {
	ID        int       `json:"id"`
	BlogID    int       `json:"blogId"`
	UserID    int       `json:"userId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func ToCommentVM(c *entity.Comment) CommentVM {
	return CommentVM{
		ID:        int(c.ID),
		BlogID:    c.BlogID,
		UserID:    c.UserID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
	}
}

func ToCommentVMs(comments []entity.Comment) []CommentVM {
	vms := make([]CommentVM, len(comments))
	for i, c := range comments {
		vms[i] = ToCommentVM(&c)
	}
	return vms
}
