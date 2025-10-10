package entity

type Blog struct {
	BaseModel
	Content  `gorm:"embedded" json:"content"` // gorm:"type:text" kalmamalıydı onu düzelttim
	Comments []Comment                        `json:"comments"` // json tag'ı "comment"ti bunu değiştirdim
	Tags     string                           `json:"tags"`
	Category string                           `json:"category"`
}
