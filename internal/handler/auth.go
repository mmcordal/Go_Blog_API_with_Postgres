package handler

import (
	"cleanArch_with_postgres/internal/service"
	"cleanArch_with_postgres/internal/viewmodel"
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	as service.AuthService
}

func NewAuthHandler(as service.AuthService) *AuthHandler {
	return &AuthHandler{as: as}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input viewmodel.RegisterRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	userRole := strings.ToLower(strings.TrimSpace(input.Role))

	resp, err := h.as.Register(context.Background(), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if userRole == "admin" {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"data":    resp,
			"message": "User created successfully! Admin rolü isteğiniz adminlere gönderildi. Admin onayına göre rolünüz belirlenecek. Not: Şu anda rolünüz 'reader' olarak kaydedildi",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data":    resp,
		"message": "User created successfully",
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input viewmodel.LoginRequest
	err := c.BodyParser(&input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	resp, err := h.as.Login(context.Background(), input.Identifier, input.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    resp,
		"message": "User login successfully!",
	})
}

func (h *AuthHandler) GetUserByUsername(c *fiber.Ctx) error {
	paramUsername := c.Params("username")
	if paramUsername == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username required"})
	}
	tokenUsername, ok := c.Locals("username").(string)
	if !ok || tokenUsername == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid token username"})
	}

	resp, err := h.as.GetUserVMByUsername(context.Background(), paramUsername, tokenUsername)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": resp,
	})
}

func (h *AuthHandler) SearchUsers(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("search"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	res, err := h.as.SearchUsers(context.Background(), q, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": res})
}

func (h *AuthHandler) UpdateUser(c *fiber.Ctx) error {
	tokenUsername, ok := c.Locals("username").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	if tokenUsername == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	tokenRole, _ := c.Locals("role").(string)

	paramUsername := c.Params("username")
	if paramUsername == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username required"})
	}
	target := paramUsername

	if tokenRole != "admin" {
		if tokenUsername != paramUsername {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not allowed"})
		}
		target = tokenUsername
	}

	var input viewmodel.UpdateRequest
	err := c.BodyParser(&input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	resp, err := h.as.UpdateUser(context.Background(), target, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    resp,
		"message": "User updated successfully!",
	})
}

func (h *AuthHandler) DeleteUser(c *fiber.Ctx) error {
	tokenUsername, ok := c.Locals("username").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	if tokenUsername == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	tokenRole, _ := c.Locals("role").(string)

	paramUsername := c.Params("username")
	if paramUsername == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username required"})
	}
	target := paramUsername

	if tokenRole != "admin" {
		if tokenUsername != paramUsername {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not allowed"})
		}
		target = tokenUsername
	}

	err := h.as.DeleteUser(context.Background(), target)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User deleted successfully!"})
}
