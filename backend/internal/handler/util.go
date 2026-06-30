package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/MaximovIlya/vk_practice_project/internal/service"
)

func toHTTPError(err error) error {
	switch err {
	case service.ErrNotFound:
		return fiber.ErrNotFound
	case service.ErrForbidden:
		return fiber.ErrForbidden
	default:
		return err
	}
}
