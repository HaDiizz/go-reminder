package controllers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HaDiizz/go-server-reminder/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func UserRegistration(db *gorm.DB, c *fiber.Ctx) error {
	var payload *models.RegisterReq

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	if len(payload.Username) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Username must be at least 3 characters."})
	}

	if len(payload.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Password must be at least 6 characters."})
	}

	if payload.Password != payload.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Passwords do not match"})

	}

	var existingUser models.User

	if err := db.Where("username = ?", payload.Username).Or("email = ?", strings.ToLower(payload.Email)).First(&existingUser).Error; err == nil {
		if existingUser.Username == payload.Username {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "error", "message": "Username already exists."})
		}
		if existingUser.Email == strings.ToLower(payload.Email) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "error", "message": "Email already exists."})
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}
	newUser := models.User{
		Username: payload.Username,
		Email:    strings.ToLower(payload.Email),
		Password: string(hashedPassword),
	}

	result := db.Create(&newUser)

	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "error", "message": "Email already exits."})
	} else if result.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": "Registration failed"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "result": fiber.Map{"user": models.FilterUserRecord(&newUser)}})
}

func UserLogin(db *gorm.DB, c *fiber.Ctx) error {
	var payload *models.LoginReq

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	if len(payload.Username) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Username must be at least 3 characters."})
	}

	if len(payload.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Password must be at least 6 characters."})
	}

	var user models.User
	result := db.First(&user, "username = ?", payload.Username)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid username or password"})
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid username or password"})
	}

	token := jwt.New(jwt.SigningMethodHS256)

	now := time.Now().UTC()
	claims := token.Claims.(jwt.MapClaims)

	claims["sub"] = user.ID
	claims["exp"] = now.Add(time.Hour * 72).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()

	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("generating JWT Token failed: %v", err)})
	}

	c.Cookie(&fiber.Cookie{
		Name:  "token",
		Value: t,
		// Path:     "/",
		// Secure:   true,
		HTTPOnly: true,
		// Domain:   "localhost",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Logged in successful", "token": t, "expiresIn": claims["exp"], "user": models.FilterUserRecord(&user)})
}

func UserLogout(c *fiber.Ctx) error {
	expired := time.Now().Add(-time.Hour * 24)
	c.Cookie(&fiber.Cookie{
		Name:    "token",
		Value:   "",
		Expires: expired,
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Logged out successful"})
}
