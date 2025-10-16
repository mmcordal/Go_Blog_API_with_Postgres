package repository

import (
	"cleanArch_with_postgres/internal/entity"
	"context"
	"fmt"
	"net/url"
	"time"

	"gorm.io/gorm"
)

type BlogRepository interface {
	Create(ctx context.Context, blog *entity.Blog) error
	Update(ctx context.Context, title string, blog *entity.Blog) error
	Delete(ctx context.Context, title string) (string, error)
	UpdateAuthorUsername(ctx context.Context, oldUsername, newUsername string) error
	GetAllTrueApproved(ctx context.Context) ([]entity.Blog, error)
	GetAllIncludeDeleted(ctx context.Context) ([]entity.Blog, error)
	GetAll(ctx context.Context) ([]entity.Blog, error)
	GetBlogsByAuthorTrueApproved(ctx context.Context, username string) ([]entity.Blog, error)
	GetBlogsByAuthorIncludeDeleted(ctx context.Context, username string) ([]entity.Blog, error)
	GetBlogsByAuthor(ctx context.Context, username string) ([]entity.Blog, error)
	GetBlogByTitleTrueApproved(ctx context.Context, title string) (*entity.Blog, error)
	GetBlogByTitle(ctx context.Context, title string) (*entity.Blog, error)
	ExistBlog(ctx context.Context, body string) (bool, error)
	SetApproval(ctx context.Context, title string, approved bool) error
	Restore(ctx context.Context, title string) error
}

type blogRepository struct {
	db *gorm.DB
}

func NewBlogRepository(db *gorm.DB) BlogRepository {
	return &blogRepository{db: db}
}

func (r *blogRepository) Create(ctx context.Context, blog *entity.Blog) error {
	err := r.db.WithContext(ctx).Create(&blog).Error
	if err != nil {
		fmt.Println("blog create error:", err)
		return err
	}
	return nil
}

func (r *blogRepository) Update(ctx context.Context, title string, blog *entity.Blog) error {
	decodedTitle, err := url.QueryUnescape(title)
	if err != nil {
		fmt.Println("title decode error:", err)
		return err
	}

	err = r.db.WithContext(ctx).Model(&entity.Blog{}).Where("title = ?", decodedTitle).
		Updates(map[string]interface{}{
			"title":      blog.Content.Title,
			"body":       blog.Content.Body,
			"type":       blog.Content.Type,
			"status":     blog.Content.Status,
			"tags":       blog.Tags,
			"category":   blog.Category,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		fmt.Println("blog update error:", err)
		return err
	}
	return nil
}

func (r *blogRepository) Delete(ctx context.Context, title string) (string, error) {
	decodedTitle, err := url.QueryUnescape(title)
	if err != nil {
		fmt.Println("title decode error:", err)
		return "", err
	}

	err = r.db.WithContext(ctx).Model(&entity.Blog{}).Where("title = ?", decodedTitle).
		Updates(map[string]interface{}{
			"deleted_at":  time.Now(),
			"status":      "deleted", // silinen blogların statüsünü "deleted" olarak güncelleniyo
			"is_approved": false,     // silinen blogun onayını kaldırıyo
		}).Error

	if err != nil {
		fmt.Println("blog delete error:", err)
		return "", err
	}
	return decodedTitle, nil
}

func (r *blogRepository) UpdateAuthorUsername(ctx context.Context, oldUsername, newUsername string) error {
	err := r.db.WithContext(ctx).Model(&entity.Blog{}).
		Where("username = ?", oldUsername).
		Updates(map[string]interface{}{
			"username":   newUsername,
			"updated_at": time.Now(),
		}).Error
	if err != nil {
		fmt.Println("blog update author username error:", err)
		return err
	}
	return nil
}

func (r *blogRepository) GetAllTrueApproved(ctx context.Context) ([]entity.Blog, error) {
	var blogs []entity.Blog

	err := r.db.WithContext(ctx).Where("is_approved = ?", true).Find(&blogs).Error
	if err != nil {
		fmt.Println("blog getAllTrueApproved error:", err)
		return nil, err
	}
	return blogs, nil
}

func (r *blogRepository) GetAllIncludeDeleted(ctx context.Context) ([]entity.Blog, error) {
	var blogs []entity.Blog
	if err := r.db.WithContext(ctx).Unscoped(). // Unscoped() GORM’un soft delete filtrelemesini kapatır
							Find(&blogs).Error; err != nil { // ve deleted_at dolu kayıtları da getirir.
		fmt.Println("blog getAllIncludeDeleted error:", err)
		return nil, err
	}
	return blogs, nil
}

func (r *blogRepository) GetAll(ctx context.Context) ([]entity.Blog, error) {
	var blogs []entity.Blog

	err := r.db.WithContext(ctx).Find(&blogs).Error
	if err != nil {
		fmt.Println("blog getAll error:", err)
		return nil, err
	}
	return blogs, nil
}

func (r *blogRepository) GetBlogsByAuthorTrueApproved(ctx context.Context, username string) ([]entity.Blog, error) {
	var blogs []entity.Blog

	err := r.db.WithContext(ctx).Where("is_approved = ?", true).
		Where("username = ?", username).Find(&blogs).Error

	if err != nil {
		fmt.Println("blog getBlogsByAuthorTrueApproved error:", err)
		return nil, err
	}
	return blogs, nil
}

func (r *blogRepository) GetBlogsByAuthorIncludeDeleted(ctx context.Context, username string) ([]entity.Blog, error) {
	var blogs []entity.Blog
	err := r.db.WithContext(ctx).
		Unscoped(). // <— soft-deleted dahil
		Where("username = ?", username).
		Find(&blogs).Error
	if err != nil {
		fmt.Println("blog getBlogsByAuthorIncludeDeleted error:", err)
		return nil, err
	}
	return blogs, nil
}

func (r *blogRepository) GetBlogsByAuthor(ctx context.Context, username string) ([]entity.Blog, error) {
	var blogs []entity.Blog

	err := r.db.WithContext(ctx).Where("username = ?", username).Find(&blogs).Error
	if err != nil {
		fmt.Println("blog getBlogsByAuthor error:", err)
		return nil, err
	}
	return blogs, nil
}

func (r *blogRepository) GetBlogByTitleTrueApproved(ctx context.Context, title string) (*entity.Blog, error) {
	var blog entity.Blog
	decodedTitle, err := url.QueryUnescape(title)
	if err != nil {
		fmt.Println("title decode error:", err)
		return nil, err
	}

	err = r.db.WithContext(ctx).Where(map[string]interface{}{"is_approved": true, "title": decodedTitle}).First(&blog).Error
	if err != nil {
		fmt.Println("blog getBlogByTitleTrueApproved error:", err)
		return nil, err
	}
	return &blog, nil
}

func (r *blogRepository) GetBlogByTitle(ctx context.Context, title string) (*entity.Blog, error) {
	var blog entity.Blog
	decodedTitle, err := url.QueryUnescape(title)
	if err != nil {
		fmt.Println("title decode error:", err)
		return nil, err
	}

	err = r.db.WithContext(ctx).Where("title = ?", decodedTitle).First(&blog).Error
	if err != nil {
		fmt.Println("blog getBlogByTitle error:", err)
		return nil, err
	}
	return &blog, nil
}

func (r *blogRepository) ExistBlog(ctx context.Context, body string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entity.Blog{}).
		Where("body = ?", body).
		Count(&count).Error

	if err != nil {
		fmt.Println("blog exist error:", err)
		return false, err
	}
	return count > 0, nil
}

func (r *blogRepository) SetApproval(ctx context.Context, title string, approved bool) error {
	decodedTitle, err := url.QueryUnescape(title)
	if err != nil {
		return err
	}

	err = r.db.WithContext(ctx).Model(&entity.Blog{}).
		Where("title = ?", decodedTitle).
		Updates(map[string]interface{}{
			"is_approved": approved,
			"updated_at":  time.Now(),
		}).Error
	if err != nil {
		fmt.Println("blog setApproval error:", err)
		return err
	}
	return nil
}

func (r *blogRepository) Restore(ctx context.Context, title string) error {
	decodedTitle, err := url.QueryUnescape(title)
	if err != nil {
		return err
	}

	tx := r.db.WithContext(ctx).
		Model(&entity.Blog{}).
		Unscoped(). // soft-deleted dahil
		Where("title = ?", decodedTitle).
		Updates(map[string]interface{}{
			"deleted_at":  nil,
			"updated_at":  time.Now(),
			"status":      "",
			"is_approved": false,
		})

	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
