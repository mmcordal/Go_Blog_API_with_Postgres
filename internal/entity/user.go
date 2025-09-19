package entity

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleReader UserRole = "reader"
	RoleWriter UserRole = "writer"
)

type User struct {
	BaseModel
	Username  string   `gorm:"type:varchar(100);unique" json:"username"`
	Email     string   `gorm:"type:varchar(100);unique" json:"email"`
	Password  string   `gorm:"type:varchar(100)" json:"-"`
	Role      UserRole `gorm:"type:varchar(100)" json:"role"`
	Followers []string `gorm:"type:varchar(100)" json:"-"`
}
