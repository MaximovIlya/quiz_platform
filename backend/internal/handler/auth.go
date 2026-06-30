package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/MaximovIlya/vk_practice_project/internal/middleware"
	"github.com/MaximovIlya/vk_practice_project/internal/service"
	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	tokens, err := h.svc.Register(c.Context(), service.RegisterParams{
		Email:    body.Email,
		Password: body.Password,
		Name:     body.Name,
	})
	if err != nil {
		switch err {
		case service.ErrEmailTaken:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		default:
			return err
		}
	}

	return c.Status(fiber.StatusCreated).JSON(tokens)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	tokens, err := h.svc.Login(c.Context(), service.LoginParams{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tokens)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	tokens, err := h.svc.Refresh(c.Context(), body.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tokens)
}

func (h *AuthHandler) SelectRole(c *fiber.Ctx) error {
	var body struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	role := db.Role(body.Role)
	if role != db.RoleORGANIZER && role != db.RolePARTICIPANT {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role"})
	}

	userID := c.Locals(middleware.CtxUserID).(string)
	if err := h.svc.SelectRole(c.Context(), userID, role); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	result, err := h.svc.ForgotPassword(c.Context(), body.Email)
	if err != nil {
		return err
	}

	// result == nil когда email не найден — возвращаем 200 чтобы не раскрывать существование аккаунта
	if result == nil {
		return c.JSON(fiber.Map{"message": "if the email exists, a reset link has been sent"})
	}

	// TODO: отправить письмо через Resend с result.Token
	return c.JSON(fiber.Map{"message": "if the email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	if err := h.svc.ResetPassword(c.Context(), body.Token, body.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "password updated"})
}
