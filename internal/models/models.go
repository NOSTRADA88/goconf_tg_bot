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
	TgID            int      `bson:"tgID"`
	Identification  string   `bson:"identification"`
	FavoriteReports []Report `bson:"favoriteReports"`
}

type Evaluation struct {
	URL         string `bson:"url"`
	TgID        int    `bson:"tgID"`
	Content     string `bson:"content"`
	Performance string `bson:"performance,omitempty"`
	Comment     string `bson:"comment,omitempty"`
}
