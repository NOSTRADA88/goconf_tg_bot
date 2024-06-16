package notificator

import (
	"fmt"
	"github.com/NOSTRADA88/telegram-bot-go/internal/config"
	"github.com/NOSTRADA88/telegram-bot-go/internal/models"
	"github.com/NOSTRADA88/telegram-bot-go/internal/storage/mongodb"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"log"
	"sync"
	"time"
)

// Notificator is a struct that contains the configuration, database, and a map of notified users.
type Notificator struct {
	Cfg           *config.Config
	Database      mongodb.DataManipulator
	NotifiedUsers map[string]bool
	mu            sync.Mutex
}

// StartNotificationScheduler starts a goroutine that periodically checks for notifications to send.
func (n *Notificator) StartNotificationScheduler(bot *gotgbot.Bot) {
	go func() {
		for {
			select {
			case <-time.After(15 * time.Second):
				err := n.notifyUpcomingReports(bot)
				if err != nil {
					fmt.Println("failed to send notification start before 10 min:", err)
				}
				err = n.notifyReportEnd(bot)
				if err != nil {
					fmt.Println("failed to send notification end of report:", err)
				}
				err = n.notifyDayEnd(bot)
				if err != nil {
					fmt.Println("failed to send notification end of day:", err)
				}
				err = n.notifyConferenceEnd(bot)
				if err != nil {
					fmt.Println("failed to send notification end of conference:", err)
				}
			}
		}
	}()
}

// notifyUpcomingReports sends a notification to users about upcoming reports.
func (n *Notificator) notifyUpcomingReports(bot *gotgbot.Bot) error {
	reports, err := n.Database.SelectReports(n.Database.Collection("report"))
	if err != nil {
		return err
	}

	users, err := n.Database.SelectUsers(n.Database.Collection("user"))
	if err != nil {
		return err
	}

	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return err
	}

	now := time.Now().In(location).Truncate(time.Second)

	n.mu.Lock()
	defer n.mu.Unlock()

	for _, report := range reports {
		startTime := report.StartTime.Truncate(time.Second)
		reportMSKTime := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(),
			startTime.Minute(), startTime.Second(), startTime.Nanosecond(), location)

		if reportMSKTime.After(now) && reportMSKTime.Before(now.Add(10*time.Minute)) {
			message := fmt.Sprintf("Доклад \"%s\" начнется меньше, чем через 10 минут в %s", report.Title, reportMSKTime.Format("15:04"))

			for _, user := range users {
				userKey := fmt.Sprintf("%d_%s", user.TgID, report.URL)
				if !n.NotifiedUsers[userKey] && (len(user.FavoriteReports) == 0 || n.isFavoriteReport(user, report.URL)) {
					n.NotifiedUsers[userKey] = true

					go func(userID int, message string) {
						msg, err := bot.SendMessage(int64(userID), message, nil)
						if err != nil {
							log.Printf("failed to send message to user %d: %v", userID, err)
							return
						}
						time.Sleep(7 * time.Second)
						_, err = bot.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
						if err != nil {
							log.Printf("failed to delete message: %v", err)
						}
					}(user.TgID, message)
				}
			}
		}
	}
	return nil
}

// notifyReportEnd sends a notification to users when a report ends.
func (n *Notificator) notifyReportEnd(bot *gotgbot.Bot) error {
	reports, err := n.Database.SelectReports(n.Database.Collection("report"))
	if err != nil {
		return err
	}

	users, err := n.Database.SelectUsers(n.Database.Collection("user"))
	if err != nil {
		return err
	}

	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return err
	}

	now := time.Now().In(location).Truncate(time.Second)

	n.mu.Lock()
	defer n.mu.Unlock()

	for _, report := range reports {
		endTime := report.StartTime.Add(time.Duration(report.Duration) * time.Minute).Truncate(time.Second)
		reportEndTime := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), endTime.Hour(),
			endTime.Minute(), endTime.Second(), endTime.Nanosecond(), location)

		if now.After(reportEndTime) {
			for _, user := range users {
				userKey := fmt.Sprintf("end_%d_%s", user.TgID, report.URL)
				if !n.NotifiedUsers[userKey] && (len(user.FavoriteReports) == 0 || n.isFavoriteReport(user, report.URL)) {
					evaluationExists, _, err := n.Database.SelectEvaluation(n.Database.Collection("evaluation"), user.TgID, report.URL)
					if err != nil {
						log.Printf("failed to check evaluation for user %d: %v", user.TgID, err)
						continue
					}
					if !evaluationExists {
						message := fmt.Sprintf("Доклад \"%s\" закончился. Пожалуйста, оцените его.", report.Title)
						msg, err := bot.SendMessage(int64(user.TgID), message, nil)
						if err != nil {
							log.Printf("failed to send message to user %d: %v", user.TgID, err)
							continue
						}
						n.NotifiedUsers[userKey] = true
						go func(msg *gotgbot.Message) {
							time.Sleep(7 * time.Second)
							_, err := bot.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
							if err != nil {
								log.Printf("failed to delete message: %v", err)
							}
						}(msg)
					}
				}
			}
		}
	}
	return nil
}

// notifyDayEnd sends a notification to users at the end of the day.
func (n *Notificator) notifyDayEnd(bot *gotgbot.Bot) error {
	reports, err := n.Database.SelectReports(n.Database.Collection("report"))
	if err != nil {
		return err
	}

	users, err := n.Database.SelectUsers(n.Database.Collection("user"))
	if err != nil {
		return err
	}

	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return err
	}

	now := time.Now().In(location).Truncate(time.Second)

	var lastReportEndTime time.Time
	for _, report := range reports {
		endTime := report.StartTime.Add(time.Duration(report.Duration) * time.Minute).Truncate(time.Second)
		reportEndTime := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), endTime.Hour(),
			endTime.Minute(), endTime.Second(), endTime.Nanosecond(), location)
		if reportEndTime.After(lastReportEndTime) {
			lastReportEndTime = reportEndTime
		}
	}

	if now.After(lastReportEndTime.Add(1 * time.Hour)) {
		for _, user := range users {
			userKey := fmt.Sprintf("day_end_%d_%s", user.TgID, lastReportEndTime.Format("02-01-2006"))
			if !n.NotifiedUsers[userKey] {
				var reportsOfDay []models.Report
				for _, report := range reports {
					if report.StartTime.Year() == lastReportEndTime.Year() && report.StartTime.YearDay() == lastReportEndTime.YearDay() {
						reportsOfDay = append(reportsOfDay, report)
					}
				}

				var unevaluatedReports []models.Report
				for _, report := range reportsOfDay {
					evaluationExists, _, err := n.Database.SelectEvaluation(n.Database.Collection("evaluation"), user.TgID, report.URL)
					if err != nil {
						log.Printf("failed to check evaluation for user %d: %v", user.TgID, err)
						continue
					}
					if !evaluationExists {
						unevaluatedReports = append(unevaluatedReports, report)
					}
				}
				if len(unevaluatedReports) > 0 {
					message := fmt.Sprintf("День закончился. Пожалуйста, оцените следующие доклады: %v", unevaluatedReports)
					msg, err := bot.SendMessage(int64(user.TgID), message, nil)
					if err != nil {
						log.Printf("failed to send message to user %d: %v", user.TgID, err)
						continue
					}
					n.NotifiedUsers[userKey] = true
					go func(msg *gotgbot.Message) {
						time.Sleep(7 * time.Second)
						_, err := bot.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
						if err != nil {
							log.Printf("failed to delete message: %v", err)
						}
					}(msg)
				}
			}
		}
	}
	return nil
}

// notifyConferenceEnd sends a notification to users two days after the conference ends.
func (n *Notificator) notifyConferenceEnd(bot *gotgbot.Bot) error {
	reports, err := n.Database.SelectReports(n.Database.Collection("report"))
	if err != nil {
		return err
	}

	users, err := n.Database.SelectUsers(n.Database.Collection("user"))
	if err != nil {
		return err
	}

	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return err
	}

	now := time.Now().In(location).Truncate(time.Second)

	conferenceEndTime := time.Date(time.Time(n.Cfg.Conference.TimeUntil).Year(), time.Time(n.Cfg.Conference.TimeUntil).Month(), time.Time(n.Cfg.Conference.TimeUntil).Day(), 23, 59, 59, 0, location)

	if now.After(conferenceEndTime.Add(2 * 24 * time.Hour)) {
		for _, user := range users {
			userKey := fmt.Sprintf("conf_end_%d_%s", user.TgID, conferenceEndTime.Format("02-01-2006"))
			if !n.NotifiedUsers[userKey] {
				var unevaluatedReports []models.Report
				for _, report := range reports {
					evaluationExists, _, err := n.Database.SelectEvaluation(n.Database.Collection("evaluation"), user.TgID, report.URL)
					if err != nil {
						log.Printf("failed to check evaluation for user %d: %v", user.TgID, err)
						continue
					}
					if !evaluationExists {
						unevaluatedReports = append(unevaluatedReports, report)
					}
				}
				if len(unevaluatedReports) > 0 {
					message := fmt.Sprintf("Конференция завершилась. Пожалуйста, оцените следующие доклады: %v", unevaluatedReports)
					msg, err := bot.SendMessage(int64(user.TgID), message, nil)
					if err != nil {
						log.Printf("failed to send message to user %d: %v", user.TgID, err)
						continue
					}
					n.NotifiedUsers[userKey] = true
					go func(msg *gotgbot.Message) {
						time.Sleep(7 * time.Second)
						_, err := bot.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
						if err != nil {
							log.Printf("failed to delete message: %v", err)
						}
					}(msg)
				}
			}
		}
	}
	return nil
}

// isFavoriteReport checks if a report is in a user's list of favorite reports.
func (n *Notificator) isFavoriteReport(user models.User, reportURL string) bool {
	for _, report := range user.FavoriteReports {
		if report.URL == reportURL {
			return true
		}
	}
	return false
}
