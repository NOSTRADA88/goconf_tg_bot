package models

import (
	"time"
)

type Report struct {
	StartTime time.Time `bson:"startTime"`
	Duration  int       `bson:"duration"`
	Title     string    `bson:"title"`
	Speakers  string    `bson:"speakers"`
	URL       string    `bson:"url"`
}

type User struct {
	ChatID          int      `bson:"chatID"`
	TgID            int      `bson:"tgID"`
	Identification  string   `bson:"identification"`
	FavoriteReports []Report `bson:"favoriteReports"`
}

type Evaluation struct {
	URL         string `bson:"url" json:"url"`
	TgID        int    `bson:"tgID" json:"tgID"`
	Content     string `bson:"content" json:"content"`
	Performance string `bson:"performance,omitempty" json:"performance"`
	Comment     string `bson:"comment,omitempty" bson:"comment"`
}
