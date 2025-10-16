// comment ile ilgili kullanıcı işlemlerini ileride ekleyeceğim.
package entity

type Comment struct {
	BaseModel
	BlogID  int    `json:"blog_id"`
	UserID  int    `json:"user_id"`
	Content string `json:"comment_content"`
}
