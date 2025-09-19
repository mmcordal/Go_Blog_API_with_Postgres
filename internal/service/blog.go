package service

import (
	"cleanArch_with_postgres/internal/entity"
	"cleanArch_with_postgres/internal/repository"
	"cleanArch_with_postgres/internal/viewmodel"
	"context"
	"errors"
	"time"
)

type BlogService interface {
	CreateBlog(ctx context.Context, blogVM *viewmodel.BlogCreateVM, username string) error
	UpdateBlog(ctx context.Context, title, username string, vm *viewmodel.BlogUpdateVM) (*viewmodel.BlogUpdateResponse, error)
	DeleteBlog(ctx context.Context, title, username string) (string, error)
	GetAllBlogs(ctx context.Context, username string) ([]viewmodel.BlogVM, error)
	GetBlogsByAuthor(ctx context.Context, paramUsername, tokenUsername string) ([]viewmodel.BlogVM, error)
	GetBlogByTitle(ctx context.Context, title, username string) (*viewmodel.BlogVM, error)
}

type blogService struct {
	br repository.BlogRepository
	ur repository.UserRepository
}

func NewBlogService(br repository.BlogRepository, ur repository.UserRepository) BlogService {
	return &blogService{br: br, ur: ur}
}

func (s *blogService) CreateBlog(ctx context.Context, blogVM *viewmodel.BlogCreateVM, username string) error {
	if username == "" {
		return errors.New("Invalid username")
	}
	user, err := s.ur.GetByUsername(ctx, username)
	if err != nil {
		return errors.New("User not found")
	}
	if user.Role != "admin" && user.Role != "writer" {
		return errors.New("User is not authorized to create a blog")
	}

	existBlog, err := s.br.ExistBlog(ctx, blogVM.Title, blogVM.Body)
	if err != nil {
		return errors.New("create blog exist error")
	}
	if existBlog {
		return errors.New("blog with the same title or body already exists")
	}
	if blogVM.Title == "" {
		return errors.New("Blog başlığı boş olamaz")
	}

	if blogVM.Type == "" {
		return errors.New("Lütfen blog tipini dolurunuz")
	}

	if blogVM.Body == "" {
		return errors.New("İçerik gövdesi boş olamaz")
	}

	if user.ID == 0 {
		return errors.New("Invalid AuthorID")
	}

	if blogVM.Status == "" {
		return errors.New("Lütfen blogun statüsünü doldurunuz")
	}

	if blogVM.Tags == "" {
		return errors.New("Blogun tag kısmı boş kalamaz")
	}

	if blogVM.Category == "" {
		return errors.New("Blogun kategorisini doğru giriniz")
	}

	blog := &entity.Blog{
		BaseModel: entity.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Content: entity.Content{
			Title:      blogVM.Title,
			Body:       blogVM.Body,
			AuthorID:   int(user.ID),
			Username:   username,
			Type:       blogVM.Type,
			IsApproved: false,
			Status:     blogVM.Status,
		},
		Tags:     blogVM.Tags,
		Category: blogVM.Category,
	}
	if user.Role == "admin" {
		blog.IsApproved = true
	}
	return s.br.Create(ctx, blog)
}

func (s *blogService) UpdateBlog(ctx context.Context, title, username string, vm *viewmodel.BlogUpdateVM) (*viewmodel.BlogUpdateResponse, error) {
	if title == "" {
		return nil, errors.New("Invalid Title")
	}

	blog, err := s.br.GetBlogByTitle(ctx, title)
	if err != nil {
		return nil, errors.New("blog not found")
	}
	user, err := s.ur.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if username != blog.Username || user.Role == "admin" { // Sadece blogun sahibi veya adminler güncelleme yapabilir
		return nil, errors.New("you are not authorized to update this blog")
	}

	if vm == nil {
		return nil, errors.New("blog is nil")
	}
	existBlog, err := s.br.ExistBlog(ctx, vm.Title, vm.Body)
	if err != nil {
		return nil, errors.New("update blog exist error")
	}
	if existBlog {
		return nil, errors.New("blog with the same title or body already exists")
	}

	if vm.Title == "" {
		return nil, errors.New("Blog başlığı boş olamaz")
	}

	if vm.Body == "" {
		return nil, errors.New("Blog içeriği boş olamaz")
	}

	blog.Content = entity.Content{
		Title:  vm.Title,
		Body:   vm.Body,
		Type:   vm.Type,
		Status: vm.Status,
	}
	blog.Tags = vm.Tags
	blog.Category = vm.Category
	blog.BaseModel.UpdatedAt = time.Now()

	resp := &viewmodel.BlogUpdateResponse{
		Username:  blog.Username,
		Title:     blog.Title,
		Body:      blog.Body,
		Type:      blog.Type,
		Tags:      blog.Tags,
		Category:  blog.Category,
		Status:    blog.Status,
		UpdatedAt: blog.UpdatedAt,
	}
	return resp, s.br.Update(ctx, title, blog)
}

func (s *blogService) DeleteBlog(ctx context.Context, title, username string) (string, error) {
	if title == "" {
		return "", errors.New("Invalid Title")
	}

	blog, err := s.br.GetBlogByTitle(ctx, title)
	if err != nil {
		return "", errors.New("blog not found")
	}
	if blog.DeletedAt.Valid {
		return "", errors.New("blog is already deleted")
	}

	user, err := s.ur.GetByUsername(ctx, username)
	if err != nil {
		return "", errors.New("user not found")
	}

	if blog.Username != username || user.Role == "admin" {
		return "", errors.New("you are not authorized to delete this blog")
	}
	return s.br.Delete(ctx, title)
}

func (s *blogService) GetAllBlogs(ctx context.Context, username string) ([]viewmodel.BlogVM, error) {
	user, err := s.ur.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user.Role != "admin" { // login olan kişi (token sahibi) admin değilse sadece onaylanmış blogları görür
		blogs, err := s.br.GetAllTrueApproved(ctx)
		if err != nil {
			return nil, errors.New("blogs get all true approved error")
		}
		return viewmodel.ToBlogVMs(blogs), nil
	}

	blogs, err := s.br.GetAll(ctx) // yukarıdaki if'e takılmayan admindir, o yüzden tüm blogları görür
	if err != nil {
		return nil, errors.New("blogs get all error")
	}
	return viewmodel.ToBlogVMs(blogs), nil
}

func (s *blogService) GetBlogsByAuthor(ctx context.Context, paramUsername, tokenUsername string) ([]viewmodel.BlogVM, error) {
	if paramUsername == "" {
		return nil, errors.New("Invalid Username")
	}
	if tokenUsername == "" {
		return nil, errors.New("Invalid Token")
	}
	user, err := s.ur.GetByUsername(ctx, tokenUsername)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if tokenUsername == paramUsername { // eğer login olan kişi (token sahibi) aratılan kullanıcının kendisiyse tüm bloglarını görebilir
		blogs, err := s.br.GetBlogsByAuthor(ctx, paramUsername)
		if err != nil {
			return nil, errors.New("blogs get by author error")
		}
		return viewmodel.ToBlogVMs(blogs), nil
	}
	if user.Role != "admin" { // login olan kişi (token sahibi) admin değilse aratılan kullanıcının sadece onaylanmış bloglarını görür
		blogs, err := s.br.GetBlogsByAuthorTrueApproved(ctx, paramUsername)
		if err != nil {
			return nil, errors.New("blogs get blogs by author true approved error")
		}
		return viewmodel.ToBlogVMs(blogs), nil
	}

	blogs, err := s.br.GetBlogsByAuthor(ctx, paramUsername) // yukarıdaki if'e takılmayan admindir, o yüzden aratılan kullanıcının tüm bloglarını görür
	if err != nil {
		return nil, errors.New("blog get by author error")
	}
	return viewmodel.ToBlogVMs(blogs), nil
}

func (s *blogService) GetBlogByTitle(ctx context.Context, title, username string) (*viewmodel.BlogVM, error) {
	if title == "" {
		return nil, errors.New("Invalid Title")
	}
	user, err := s.ur.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("user not found")
	}
	blog, err := s.br.GetBlogByTitle(ctx, title)
	if err != nil {
		return nil, errors.New("blog not found")
	}
	if blog.Username != username { // login olan kişi (token sahibi) çağırılan blogun yazarıysa onaylanmasa bile görüntülesin
		return viewmodel.ToBlogVM(blog), nil
	}
	if user.Role != "admin" { // login olan kişi (token sahibi) admin değilse sadece onaylanmış bir blogu aratıp görebilir
		blog, err := s.br.GetBlogByTitleTrueApproved(ctx, title)
		if err != nil {
			return nil, errors.New("blogs get all true approved error")
		}
		return viewmodel.ToBlogVM(blog), nil
	}
	return viewmodel.ToBlogVM(blog), nil
}
