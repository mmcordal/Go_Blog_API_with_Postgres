package handler

import (
	"cleanArch_with_postgres/internal/service"
	"cleanArch_with_postgres/internal/viewmodel"
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type BlogHandler struct {
	bs service.BlogService
}

func NewBlogHandler(bs service.BlogService) *BlogHandler {
	return &BlogHandler{bs: bs}
}

func (h *BlogHandler) CreateBlog(c *fiber.Ctx) error {
	var input viewmodel.BlogCreateVM
	err := c.BodyParser(&input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid input json",
			"message": err.Error(),
		})
	}

	username, ok := c.Locals("username").(string)
	if !ok || username == "" { // ok la username ayır
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid token username",
		})
	}
	/*
		userID, ok := c.Locals("user_id").(uint)
		if !ok || userID == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid token userID",
			})
		}
	*/

	err = h.bs.CreateBlog(context.Background(), &input, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Blog created successfully",
		"username": username,
		//"userID":   userID,
	})
}

func (h *BlogHandler) UpdateBlog(c *fiber.Ctx) error {
	title := c.Params("title") // param title
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "title required",
		})
	}

	username, ok := c.Locals("username").(string) // token username
	if !ok || username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid token username",
		})
	}

	var input viewmodel.BlogUpdateVM
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid input json",
			"message": err.Error(),
		})
	}

	resp, err := h.bs.UpdateBlog(context.Background(), title, username, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    resp,
		"message": "Blog updated successfully",
	})
}

func (h *BlogHandler) DeleteBlog(c *fiber.Ctx) error {
	title := c.Params("title") // param title		/	undecodedTitle
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid title param",
		})
	}
	username, ok := c.Locals("username").(string)
	if !ok || username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid token username",
		})
	}
	decodedTitle, err := h.bs.DeleteBlog(context.Background(), title, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "Blog deleted successfully",
		"title":       decodedTitle,
		"blog_author": username,
	})
}

func (h *BlogHandler) GetAllBlogs(c *fiber.Ctx) error {
	username, ok := c.Locals("username").(string) // token username
	if !ok || username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid token username"})
	}

	includeDeleted := false
	// include_deleted=true mu?
	if v := c.Query("include_deleted"); v == "true" || v == "1" {
		// rol kontrolü: sadece admin kullanabilsin
		if role, _ := c.Locals("role").(string); role == "admin" {
			includeDeleted = true
		}
	}

	resp, err := h.bs.GetAllBlogsWithOptions(context.Background(), username, includeDeleted)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": resp})
}

func (h *BlogHandler) GetBlogsByAuthor(c *fiber.Ctx) error {
	paramUsername := c.Params("username") // param username
	if paramUsername == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid param username"})
	}
	tokenUsername, ok := c.Locals("username").(string)
	if !ok || tokenUsername == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid token username"})
	}

	// include_deleted=true/1/yes destekle
	inc := strings.ToLower(c.Query("include_deleted"))
	includeDeleted := inc == "1" || inc == "true" || inc == "yes"

	resp, err := h.bs.GetBlogsByAuthor(context.Background(), paramUsername, tokenUsername, includeDeleted)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Blogs Author: " + paramUsername,
		"data":    resp,
	})
}

func (h *BlogHandler) GetBlogsByAuthorIncludeDeleted(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid username"})
	}

	resp, err := h.bs.GetBlogsByAuthorIncludeDeleted(context.Background(), username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": resp,
	})
}

func (h *BlogHandler) GetBlogByTitle(c *fiber.Ctx) error {
	title := c.Params("title")
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid title"})
	}
	username, ok := c.Locals("username").(string)
	if !ok || username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid token username"})
	}
	resp, err := h.bs.GetBlogByTitle(context.Background(), title, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": resp})
}

func (h *BlogHandler) ApproveBlog(c *fiber.Ctx) error {
	title := c.Params("title")
	username, _ := c.Locals("username").(string)
	if err := h.bs.ApproveBlog(context.Background(), title, username, true); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "approved"})
}

func (h *BlogHandler) UnapproveBlog(c *fiber.Ctx) error {
	title := c.Params("title")
	username, _ := c.Locals("username").(string)
	if err := h.bs.ApproveBlog(context.Background(), title, username, false); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "unapproved"})
}

func (h *BlogHandler) RestoreBlog(c *fiber.Ctx) error {
	title := c.Params("title")
	username, _ := c.Locals("username").(string)
	if err := h.bs.RestoreBlog(context.Background(), title, username); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "restored"})
}
