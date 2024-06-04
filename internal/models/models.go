// Package models provides the data structures used in the application.
package models

import (
	"time"
)

// Report represents a report with its start time, duration, title, speakers, and URL.
type Report struct {
	StartTime time.Time `bson:"startTime"` // StartTime is the start time of the report.
	Duration  int       `bson:"duration"`  // Duration is the duration of the report in minutes.
	Title     string    `bson:"title"`     // Title is the title of the report.
	Speakers  string    `bson:"speakers"`  // Speakers is a string of speakers' names.
	URL       string    `bson:"url"`       // URL is the URL of the report.
}

// User represents a user with their chat ID, Telegram ID, identification, and favorite reports.
type User struct {
	ChatID          int      `bson:"chatID"`          // ChatID is the ID of the chat with the user.
	TgID            int      `bson:"tgID"`            // TgID is the Telegram ID of the user.
	Identification  string   `bson:"identification"`  // Identification is the identification of the user.
	FavoriteReports []Report `bson:"favoriteReports"` // FavoriteReports is a slice of the user's favorite reports.
}

// Evaluation represents an evaluation with its URL, Telegram ID, content, performance, and comment.
type Evaluation struct {
	URL         string `bson:"url" json:"url"`                           // URL is the URL of the evaluation.
	TgID        int    `bson:"tgID" json:"tgID"`                         // TgID is the Telegram ID of the user who made the evaluation.
	Content     string `bson:"content" json:"content"`                   // Content is the content of the evaluation.
	Performance string `bson:"performance,omitempty" json:"performance"` // Performance is the performance rating of the evaluation. It is optional.
	Comment     string `bson:"comment,omitempty" bson:"comment"`         // Comment is the comment of the evaluation. It is optional.
}
