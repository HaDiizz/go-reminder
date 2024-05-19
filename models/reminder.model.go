package models

import (
	"time"

	"gorm.io/gorm"
)

type Reminder struct {
	gorm.Model
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"not null" json:"description"`
	RemindAt    time.Time `gorm:"not null" json:"remindAt"`
	UserID      uint      `gorm:"not null" json:"userId"`
	User        User
}

type ReminderReq struct {
	Title       string `json:"title"  validate:"required"`
	Description string `json:"description"  validate:"required"`
	RemindAt    string `json:"remindAt"  validate:"required"`
	UserID      uint   `json:"userId"  validate:"required"`
}

type ReminderResponse struct {
	ID          uint         `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	RemindAt    time.Time    `json:"remindAt"`
	UserID      uint         `json:"userId"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	User        UserResponse `json:"user"`
}

func FilterReminderRecord(reminder *Reminder) ReminderResponse {
	return ReminderResponse{
		ID:          reminder.ID,
		Title:       reminder.Title,
		Description: reminder.Description,
		RemindAt:    reminder.RemindAt,
		UserID:      reminder.UserID,
		CreatedAt:   reminder.CreatedAt,
		UpdatedAt:   reminder.UpdatedAt,
		User:        FilterUserRecord(&reminder.User),
	}
}
