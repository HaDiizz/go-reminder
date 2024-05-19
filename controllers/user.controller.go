package controllers

import (
	"github.com/HaDiizz/go-server-reminder/models"
	"github.com/gofiber/fiber/v2"
)

func UserInfo(c *fiber.Ctx) error {
	userInfo := c.Locals("userInfo").(models.UserResponse)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "result": fiber.Map{"userInfo": userInfo}})
}
