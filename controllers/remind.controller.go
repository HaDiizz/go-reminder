package controllers

import (
	"strconv"
	"time"

	"github.com/HaDiizz/go-server-reminder/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateRemind(db *gorm.DB, c *fiber.Ctx) error {
	var payload *models.ReminderReq

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	if len(payload.Title) < 1 || len(payload.Title) > 30 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Title must be between 1 and 30 characters."})
	}
	if len(payload.Description) < 1 || len(payload.Description) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Description must be between 1 and 100 characters."})
	}

	userInfo, ok := c.Locals("userInfo").(models.UserResponse)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Unauthorized access."})
	}

	layout := time.RFC3339
	remindAt, err := time.Parse(layout, payload.RemindAt)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid reminder time format."})
	}

	if remindAt.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Reminder time must be in the future."})
	}

	reminder := models.Reminder{
		Title:       payload.Title,
		Description: payload.Description,
		RemindAt:    remindAt,
		UserID:      userInfo.ID,
	}

	if err := db.Create(&reminder).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "result": fiber.Map{"reminder": reminder}})
}

func GetReminders(db *gorm.DB, c *fiber.Ctx) error {
	queryValue := c.Query("tab")
	userInfo := c.Locals("userInfo").(models.UserResponse)

	var reminders []models.Reminder
	var err error

	if queryValue == "active" {
		err = db.Preload("User").Where("remind_at > ? AND user_id = ?", time.Now(), userInfo.ID).Find(&reminders).Error
	} else if queryValue == "inactive" {
		err = db.Preload("User").Where("remind_at < ? AND user_id = ?", time.Now(), userInfo.ID).Find(&reminders).Error
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid tab value"})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Database error", "error": err.Error()})
	}
	reminderResponses := make([]models.ReminderResponse, len(reminders))
	for i, reminder := range reminders {
		reminderResponses[i] = models.FilterReminderRecord(&reminder)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "result": fiber.Map{"reminders": reminderResponses}})
}

func DeleteReminder(db *gorm.DB, c *fiber.Ctx) error {
	reminderId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid reminder ID."})
	}

	userInfo, ok := c.Locals("userInfo").(models.UserResponse)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Unauthorized access."})
	}

	var reminder models.Reminder

	if err := db.First(&reminder, reminderId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Reminder not found."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Database error.", "error": err.Error()})
	}

	if userInfo.ID != reminder.UserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "error", "message": "You are not authorized to delete this reminder."})
	}

	if err := db.First(&reminder, reminderId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Reminder not found."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Database error.", "error": err.Error()})
	}

	if err := db.Delete(&reminder).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to delete reminder.", "error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Reminder deleted successfully."})
}
