package router

import (
	"cleanArch_with_postgres/internal/handler"
	"cleanArch_with_postgres/internal/infrastructure/app"
	"cleanArch_with_postgres/internal/middleware"
	"cleanArch_with_postgres/internal/repository"
	"cleanArch_with_postgres/internal/service"
)

type Router struct{}

func NewRouter() *Router {
	return &Router{}
}

func (Router) RegisterRouter(a *app.App) {

	app := a.FiberApp
	db := a.DB

	// Repositories
	ur := repository.NewUserRepository(db)
	br := repository.NewBlogRepository(db)

	// Services
	as := service.NewAuthService(ur)
	bs := service.NewBlogService(br, ur)

	// Handlers
	ah := handler.NewAuthHandler(as)
	bh := handler.NewBlogHandler(bs)

	v1 := app.Group("/api/v1")

	v1.Post("/register", ah.Register)
	v1.Post("/login", ah.Login)

	v1.Use(middleware.JWTMiddleware())

	// Auth
	v1.Get("/user/:username", ah.GetUserByUsername)
	v1.Put("/user/:username", ah.UpdateUser)
	v1.Delete("/user/:username", ah.DeleteUser)

	// Blog
	v1.Get("/blogs", bh.GetAllBlogs)
	v1.Get("/blogs/:username", bh.GetBlogsByAuthor)
	v1.Get("/blog/:title", bh.GetBlogByTitle)
	v1.Post("/blog", bh.CreateBlog)
	v1.Put("/blog/:title", bh.UpdateBlog)
	v1.Delete("/blog/:title", bh.DeleteBlog)
}
