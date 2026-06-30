package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/MaximovIlya/vk_practice_project/internal/middleware"
	"github.com/MaximovIlya/vk_practice_project/internal/service"
)

type SessionHandler struct {
	svc *service.SessionService
}

func NewSessionHandler(svc *service.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

func (h *SessionHandler) GetOrCreate(c *fiber.Ctx) error {
	userID := c.Locals(middleware.CtxUserID).(string)

	sess, err := h.svc.GetOrCreate(c.Context(), c.Params("id"), userID)
	if err != nil {
		return toHTTPError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(mapSessionWithPlayers(sess.QuizSession, sess.Players))
}

func (h *SessionHandler) Get(c *fiber.Ctx) error {
	userID := c.Locals(middleware.CtxUserID).(string)

	sess, err := h.svc.GetLatest(c.Context(), c.Params("id"), userID)
	if err != nil {
		return toHTTPError(err)
	}

	return c.JSON(mapSessionWithPlayers(sess.QuizSession, sess.Players))
}

func (h *SessionHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals(middleware.CtxUserID).(string)

	if err := h.svc.Delete(c.Context(), c.Params("id"), userID); err != nil {
		return toHTTPError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
