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
	GetAllTrueApproved(ctx context.Context) ([]entity.Blog, error)
	GetAll(ctx context.Context) ([]entity.Blog, error)
	GetBlogsByAuthorTrueApproved(ctx context.Context, username string) ([]entity.Blog, error)
	GetBlogsByAuthor(ctx context.Context, username string) ([]entity.Blog, error)
	GetBlogByTitleTrueApproved(ctx context.Context, title string) (*entity.Blog, error)
	GetBlogByTitle(ctx context.Context, title string) (*entity.Blog, error)
	ExistBlog(ctx context.Context, title, body string) (bool, error)
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
			"deleted_at": time.Now(),
		}).Error

	if err != nil {
		fmt.Println("blog delete error:", err)
		return "", err
	}
	return decodedTitle, nil
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

	err = r.db.WithContext(ctx).Where("title", decodedTitle).First(&blog).Error
	if err != nil {
		fmt.Println("blog getBlogByTitle error:", err)
		return nil, err
	}
	return &blog, nil
}

func (r *blogRepository) ExistBlog(ctx context.Context, title, body string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entity.Blog{}).
		Where("title = ? OR body = ?", title, body).
		Count(&count).Error

	if err != nil {
		fmt.Println("blog exist error:", err)
		return false, err
	}
	return count > 0, nil
}
