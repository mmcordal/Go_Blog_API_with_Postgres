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

	input.Role = strings.ToLower(strings.TrimSpace(input.Role))

	resp, err := h.as.Register(context.Background(), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if input.Role == "admin" {
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
	if err := c.BodyParser(&input); err != nil {
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
	paramUsername := strings.TrimSpace(c.Params("username"))
	if paramUsername == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username required"})
	}
	tokenUsername, _ := c.Locals("username").(string)
	if tokenUsername == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	resp, err := h.as.GetUserVMByUsername(context.Background(), paramUsername, tokenUsername)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": resp})
}

func (h *AuthHandler) SearchUsers(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("search"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	includeDeleted := false
	if v := c.Query("include_deleted"); v == "true" || v == "1" {
		includeDeleted = true
	}

	viewerUsername, _ := c.Locals("username").(string) // token’daki kullanıcı adı

	res, err := h.as.SearchUsersWithOptions(context.Background(), viewerUsername, q, limit, includeDeleted)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": res})
}

func (h *AuthHandler) RestoreUser(c *fiber.Ctx) error {
	// sadece admin
	role, _ := c.Locals("role").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not allowed"})
	}

	username := strings.TrimSpace(c.Params("username"))
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username required"})
	}

	if err := h.as.RestoreUser(context.Background(), username); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User restored successfully!"})
}

func (h *AuthHandler) UpdateUser(c *fiber.Ctx) error {
	tokenUsername, _ := c.Locals("username").(string)
	if tokenUsername == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	tokenRole, _ := c.Locals("role").(string)

	paramUsername := strings.TrimSpace(c.Params("username"))
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
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	resp, err := h.as.UpdateUser(context.Background(), target, &input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": resp, "message": "User updated successfully!"})
}

func (h *AuthHandler) DeleteUser(c *fiber.Ctx) error {
	tokenUsername, _ := c.Locals("username").(string)
	if tokenUsername == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	tokenRole, _ := c.Locals("role").(string)

	paramUsername := strings.TrimSpace(c.Params("username"))
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

	if err := h.as.DeleteUser(context.Background(), target); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User deleted successfully!"})
}

// ---------- ROLE REQUESTS ----------

func (h *AuthHandler) RequestAdminRole(c *fiber.Ctx) error {
	tokenUsername, _ := c.Locals("username").(string)
	if tokenUsername == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.BodyParser(&body)

	vm, err := h.as.RequestAdminRole(context.Background(), tokenUsername, body.Reason)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": vm, "message": "Talebiniz alındı"})
}

func (h *AuthHandler) ListRoleRequests(c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not allowed"})
	}
	status := c.Query("status")
	limitStr := c.Query("limit", "100")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 100
	}

	list, err := h.as.ListRoleRequests(context.Background(), status, limit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": list})
}

func (h *AuthHandler) ApproveRoleRequest(c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	admin, _ := c.Locals("username").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not allowed"})
	}
	id64, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id64 == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.as.ApproveRoleRequest(context.Background(), uint(id64), admin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Talep onaylandı"})
}

func (h *AuthHandler) RejectRoleRequest(c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	admin, _ := c.Locals("username").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not allowed"})
	}
	id64, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id64 == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.as.RejectRoleRequest(context.Background(), uint(id64), admin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Talep reddedildi"})
}

// ---------- ME ENDPOİNTLERİ ----------

func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	if username == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	vm, err := h.as.GetUserVMByUsername(context.Background(), username, username)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": vm})
}

func (h *AuthHandler) UpdateMe(c *fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	if username == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var in viewmodel.UpdateRequest
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}
	resp, err := h.as.UpdateUser(context.Background(), username, &in)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": resp, "message": "User updated successfully!"})
}

func (h *AuthHandler) DeleteMe(c *fiber.Ctx) error {
	username, _ := c.Locals("username").(string)
	if username == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	if err := h.as.DeleteUser(context.Background(), username); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User deleted successfully!"})
}
