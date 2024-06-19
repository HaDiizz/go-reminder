package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/HaDiizz/go-server-reminder/controllers"
	"github.com/HaDiizz/go-server-reminder/middleware"
	"github.com/HaDiizz/go-server-reminder/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_DB_HOST"), os.Getenv("POSTGRES_DB_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("failed to connect to database")
	}

	db.AutoMigrate(&models.User{}, &models.Reminder{})

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://thereminder.vercel.app",
		AllowCredentials: true,
	}))

	routes := app.Group("/api")

	routes.Post("/register", func(c *fiber.Ctx) error {
		return controllers.UserRegistration(db, c)
	})

	routes.Post("/login", func(c *fiber.Ctx) error {
		return controllers.UserLogin(db, c)
	})

	routes.Post("/logout", func(c *fiber.Ctx) error {
		return controllers.UserLogout(c)
	})

	routes.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello world",
		},
		)
	})

	routes.Use(func(c *fiber.Ctx) error {
		return middleware.Auth(db, c)
	})

	routes.Get("/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello",
		},
		)
	})

	routes.Get("/userInfo", func(c *fiber.Ctx) error {
		return controllers.UserInfo(c)
	})
	routes.Get("/reminders", func(c *fiber.Ctx) error {
		return controllers.GetReminders(db, c)
	})

	routes.Post("/reminders/create", func(c *fiber.Ctx) error {
		return controllers.CreateRemind(db, c)
	})

	routes.Delete("/reminders/delete/:id", func(c *fiber.Ctx) error {
		return controllers.DeleteReminder(db, c)
	})

	log.Fatal(app.Listen("localhost:8080"))
}
