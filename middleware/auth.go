package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/HaDiizz/go-server-reminder/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

func Auth(db *gorm.DB, c *fiber.Ctx) error {
	var cookie string
	authorization := c.Get("Authorization")

	if strings.HasPrefix(authorization, "Bearer ") {
		cookie = strings.TrimPrefix(authorization, "Bearer ")
	} else if c.Cookies("token") != "" {
		cookie = c.Cookies("token")
	}

	if cookie == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Unauthorized"})
	}

	token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})

	if err != nil || !token.Valid {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	claim := token.Claims.(jwt.MapClaims)

	var user models.User
	db.First(&user, "id = ?", fmt.Sprint(claim["sub"]))

	if float64(user.ID) != float64(claim["sub"].(float64)) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "error", "message": "Forbidden"})
	}

	c.Locals("userInfo", models.FilterUserRecord(&user))

	return c.Next()
}
