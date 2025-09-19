package entity

type Blog struct {
	BaseModel
	Content  `gorm:"type:text" json:"content"`
	Comments []Comment `json:"comment"`
	Tags     string    `json:"tags"`
	Category string    `json:"category"`
}
