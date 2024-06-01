package handlers

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/NOSTRADA88/telegram-bot-go/internal/models"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	html = "html"
)

func (c *Client) startHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	var err error

	state, err := c.FSM.GetState(context.Background(), ctx.EffectiveUser.Id)

	if err != nil {
		return err
	}

	switch state {

	case "":

		if err = c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, start); err != nil {
			return err
		}

		_, err = bot.SendMessage(ctx.EffectiveChat.Id, "👋 Здравствуйте, перед началом использования бота, введите, пожалуйста, ваш билет/почту/ФИО (одно на выбор). Эта информация требуется для вашей идентификации 👤\n\nНе переживайте, вы сможете изменить её в любой момент 😊", nil)

	case uploadSchedule:

		if err = c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, menu); err != nil {
			return err
		}

		_, err = bot.SendMessage(ctx.Message.Chat.Id,
			"Спасибо, что вы загрузили расписание. Так держать, скушайте печеньку!",
			&gotgbot.SendMessageOpts{
				ParseMode:   html,
				ReplyMarkup: mainMenuAdminKB(),
			})

		if err != nil {
			return err
		}

	case changeIdentification, confInfo:

		errS := c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, menu)

		if errS != nil {
			return errS
		}

		user, errS := c.Database.SelectUser(c.Database.Collection("user"), int(ctx.EffectiveUser.Id))

		if errS != nil {
			return errS
		}

		if _, exists := c.Cfg.Administrators.IDsInMap[int(ctx.Message.From.Id)]; exists {

			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				"Вижу, что вы недавно изменили свою идентификацию. Не забывайте оставлять отзывы о просмотренных докладах. Обратная связь крайне важна",
				&gotgbot.SendMessageOpts{
					ParseMode:   html,
					ReplyMarkup: mainMenuAdminKB(),
				})

			if err != nil {
				return err
			}

		} else {

			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				fmt.Sprintf("Привет %s, я @%s. Вижу, что вы недавно изменили свою идентификацию. Не забывайте оставлять отзывы о докладах", user.Identification, bot.User.Username),
				&gotgbot.SendMessageOpts{
					ParseMode:   html,
					ReplyMarkup: mainMenuUserKB(),
				})

			if err != nil {
				return err
			}

		}
	case start:
		if strings.HasPrefix(ctx.EffectiveMessage.Text, "/") {

			_, errS := bot.SendMessage(ctx.EffectiveChat.Id, "Простите, но имя не может начинаться с \"/\". Введите, что-нибудь другое... билет... ФИО или вашу почту.", nil)

			if errS != nil {
				return errS
			}

			return nil
		}
	default:
		user, errS := c.Database.SelectUser(c.Database.Collection("user"), int(ctx.EffectiveUser.Id))

		if errS != nil {
			return errS
		}

		if _, exists := c.Cfg.Administrators.IDsInMap[int(ctx.Message.From.Id)]; exists {
			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				fmt.Sprintf("Снова здравствуй, уважаемый %s. Честно, сам не знаю, чтобы я хотел, будь я тобой. Попробуй понажимать на другие кнопки что ли...", user.Identification),
				&gotgbot.SendMessageOpts{
					ParseMode:   html,
					ReplyMarkup: mainMenuAdminKB(),
				})

			if err != nil {
				return err
			}

		} else {
			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				fmt.Sprintf("Снова здравствуй, уважаемый %s. Честно, сам не знаю, чтобы я хотел, будь я тобой. Попробуй понажимать на другие кнопки что ли...", user.Identification, bot.User.Username),
				&gotgbot.SendMessageOpts{
					ParseMode:   html,
					ReplyMarkup: mainMenuUserKB(),
				})

			if err != nil {
				return err
			}
		}
	}

	return nil

}

func (c *Client) textHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	state, err := c.FSM.GetState(context.Background(), ctx.EffectiveUser.Id)

	if err != nil {
		return err
	}

	switch state {

	case start:
		if strings.HasPrefix(ctx.EffectiveMessage.Text, "/") {

			_, errS := bot.SendMessage(ctx.EffectiveChat.Id, "Простите, но имя не может начинаться с \"/\". Введите, что-нибудь другое... билет... ФИО или вашу почту.",
				nil)

			if errS != nil {
				return errS
			}

			return nil
		}

		coll := c.Database.Collection("user")

		errI := c.Database.InsertOne(coll, models.User{
			TgID: int(ctx.EffectiveUser.Id), Identification: ctx.EffectiveMessage.Text, FavoriteReports: []models.Report{}})

		if errI != nil {
			return errI
		}

		if err = c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, menu); err != nil {
			return err
		}

		user, errS := c.Database.SelectUser(coll, int(ctx.EffectiveUser.Id))

		if errS != nil {
			return errS
		}

		if _, exists := c.Cfg.Administrators.IDsInMap[int(ctx.Message.From.Id)]; exists {
			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				fmt.Sprintf("Привет %s, я @%s. А вы знали, что у вас есть власть, которая и снилась обычным пользователям этого бота? Да? Тогда загрузите уже расписание с докладами!", user.Identification, bot.User.Username),
				&gotgbot.SendMessageOpts{
					ParseMode:   html,
					ReplyMarkup: mainMenuAdminKB(),
				})

			if err != nil {
				return err
			}

		} else {
			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				fmt.Sprintf("Добро пожаловать %s, я @%s. Рекомендую ознакомиться с предстоящими докладами в \"👀 Посмотреть доклады\". Я точно уверен, что ты найдёшь что-то для себя. Также вся навигации осущствляется кнопкапки ниже", user.Identification, bot.User.Username),
				&gotgbot.SendMessageOpts{
					ParseMode:   html,
					ReplyMarkup: mainMenuUserKB(),
				})

			if err != nil {
				return err
			}

		}
	case menu:
		user, errS := c.Database.SelectUser(c.Database.Collection("user"), int(ctx.EffectiveUser.Id))

		if errS != nil {
			return errS
		}

		if _, exists := c.Cfg.Administrators.IDsInMap[int(ctx.Message.From.Id)]; exists {
			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				fmt.Sprintf("%s, что привело вас в главное меню? Хотя мне без разницы... Просто нажмите любую кнопку", user.Identification),
				&gotgbot.SendMessageOpts{
					ParseMode:   html,
					ReplyMarkup: mainMenuAdminKB(),
				})

			if err != nil {
				return err
			}

		} else {
			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				fmt.Sprintf("Здравствуйте %s. Что желаете ? К сожалению, я могу предложить вам только кнопочки ниже...", user.Identification),
				&gotgbot.SendMessageOpts{
					ParseMode:   html,
					ReplyMarkup: mainMenuUserKB(),
				})

			if err != nil {
				return err
			}

		}
	case changeIdentification:

		if strings.HasPrefix(ctx.EffectiveMessage.Text, "/") {

			_, errS := bot.SendMessage(ctx.EffectiveChat.Id, "Простите, но имя не может начинаться с \"/\". Введите, что-нибудь другое... билет... ФИО или вашу почту.",
				&gotgbot.SendMessageOpts{ParseMode: html, ReplyMarkup: backToMainMenuKB()})

			if errS != nil {
				return errS
			}

			return nil
		}

		coll := c.Database.Collection("user")

		ok, errU := c.Database.UpdateUserID(coll, int(ctx.EffectiveUser.Id), ctx.EffectiveMessage.Text)

		if errU != nil {
			return errU
		}

		if ok {
			user, errS := c.Database.SelectUser(coll, int(ctx.EffectiveUser.Id))
			if errS != nil {
				return errS
			}
			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				fmt.Sprintf("Я поменял на %s. Если вы снова хотите изменить вашу идентификацию, пожалуйста, отправьте сообщение в этот же чат. В ином случае, нажмите \"⬅️ Назад\" или /start", user.Identification),
				&gotgbot.SendMessageOpts{ParseMode: html, ReplyMarkup: backToMainMenuKB()})
			if err != nil {
				return err
			}
		}
	case uploadSchedule, viewReports:
		_, errD := bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)

		if errD != nil {
			return errD
		}
	}
	if len(strings.Split(state, ";")) == 5 {

		stateSeparated := strings.Split(state, ";")
		text := ctx.EffectiveMessage.Text
		evaluation := models.Evaluation{URL: stateSeparated[1], TgID: int(ctx.Message.From.Id),
			Content: stateSeparated[len(stateSeparated)-2], Performance: stateSeparated[len(stateSeparated)-1],
			Comment: text}

		if err := c.Database.InsertOne(c.Database.Collection("evaluation"), evaluation); err != nil {
			return err
		}

		_, err := bot.SendMessage(ctx.EffectiveChat.Id, "Ваш отзыв успешно добавлен!", &gotgbot.SendMessageOpts{
			ReplyMarkup: afterMarkKB(),
			ParseMode:   html,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) confInfoCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	var err error

	err = c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, confInfo)
	if err != nil {
		return err
	}

	cb := ctx.Update.CallbackQuery

	_, _, err = cb.Message.EditText(bot, fmt.Sprintf("📅 О конференции\n\n🎉 Название: %s\n🌐 Сайт: %s\n\n🕒 Время начала: %s\n🕙 Время окончания: %s\n\n",
		c.Cfg.Conference.Name, c.Cfg.Conference.URL, time.Time(c.Cfg.Conference.TimeFrom).Format("02.01.2006 15:04"),
		time.Time(c.Cfg.Conference.TimeUntil).Format("02.01.2006 15:04")),
		&gotgbot.EditMessageTextOpts{ParseMode: html, ReplyMarkup: backToMainMenuKB()})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) backCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	var err error

	err = c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, menu)

	if err != nil {
		return err
	}

	cb := ctx.Update.CallbackQuery

	if _, exists := c.Cfg.Administrators.IDsInMap[int(cb.From.Id)]; exists {
		_, _, err = cb.Message.EditText(bot,
			"Вы вернулись в главное меню, как удобно... что я контролирую ваши нажатия. ",
			&gotgbot.EditMessageTextOpts{ParseMode: html, ReplyMarkup: mainMenuAdminKB()})

		if err != nil {
			return err
		}

	} else {
		_, _, err = cb.Message.EditText(bot,
			"Вы вернулись в главное меню, как удобно... что я контролирую ваши нажатия. ",
			&gotgbot.EditMessageTextOpts{ParseMode: html, ReplyMarkup: mainMenuUserKB()})

		if err != nil {
			return err
		}

	}

	return nil
}

func (c *Client) uploadScheduleCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	var err error

	err = c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, uploadSchedule)

	if err != nil {
		return err
	}

	cb := ctx.Update.CallbackQuery

	_, _, err = cb.Message.EditText(bot, "Загрузите файл с расписанием: ", &gotgbot.EditMessageTextOpts{
		ParseMode:   html,
		ReplyMarkup: backToMainMenuKB(),
	})

	if err != nil {
		return err
	}

	return nil

}

func getFormatReports(data []models.Report) string {
	var reports string

	for ind, report := range data {
		reports += fmt.Sprintf("%v. %v время начала: %v\n\n%s - %s\n\n", ind+1, report.StartTime.Format("02.01.2006"), report.StartTime.Format("15:04"), report.Speakers, report.Title)
	}

	return reports
}

func (c *Client) fileHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	state, err := c.FSM.GetState(context.Background(), ctx.EffectiveUser.Id)

	if err != nil {
		return err
	}

	switch state {

	case uploadSchedule:

		fileExtension := strings.Split(ctx.EffectiveMessage.Document.FileName, ".")[len(strings.Split(ctx.EffectiveMessage.Document.FileName, "."))-1]

		if fileExtension != "scv" {
			_, errSF := bot.SendMessage(ctx.EffectiveChat.Id, "Простите, но я работают исключительно с файлами в формате .csv", nil)

			if errSF != nil {
				return errSF
			}

			return fmt.Errorf("incorrect file extension: expected \"scv\" got \"%v\"", fileExtension)
		}

		file, errF := bot.GetFile(ctx.EffectiveMessage.Document.FileId, nil)

		if errF != nil {
			return errF
		}

		response, errG := http.Get(fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.Cfg.Telegram.Token, file.FilePath))

		if errG != nil {
			return errG
		}

		defer func() {
			if errC := response.Body.Close(); errC != nil {
			}
		}()

		scanner := bufio.NewScanner(response.Body)

		var reports []interface{}

		for scanner.Scan() {

			reader := csv.NewReader(strings.NewReader(scanner.Text()))

			reader.TrimLeadingSpace = true

			record, errR := reader.Read()

			if errR != nil {
				return errR
			}

			if len(record) != 5 {

				_, errSL := bot.SendMessage(ctx.EffectiveChat.Id, "Длина каждой строки в файле должна быть равна 5!\n```\nStart (MSK Time Zone),Duration (min),Title,Speakers,URL\n```", nil)

				if errSL != nil {
					return errSL
				}

				return fmt.Errorf("invalid line length: excpeted 5, got %v", len(record))
			}

			t, errT := time.Parse("02/01/2006 15:04:05", record[0])

			if errT != nil {
				_, errST := bot.SendMessage(ctx.EffectiveChat.Id, "Строка содержит неправильный формат времени!\n\n02/01/2006 15:04:05\n\n", nil)

				if errST != nil {
					return errST
				}

				return errT
			}

			if t.Before(time.Time(c.Cfg.Conference.TimeFrom)) || time.Time(c.Cfg.Conference.TimeUntil).Before(t) {
				_, errST := bot.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("Время доклада \"%s\" не попадает в интервал с %s по %s", record[2], time.Time(c.Cfg.Conference.TimeFrom).Format("02.01.2006 15:04:05"), time.Time(c.Cfg.Conference.TimeUntil).Format("02.01.2006 15:04:05")), nil)

				if errST != nil {
					return errST
				}

				return fmt.Errorf("mismatched time interval")
			}

			var duration int

			_, errS := fmt.Sscanf(record[1], "%d", &duration)

			if errS != nil {
				return errS
			}

			report := models.Report{StartTime: t, Duration: duration, Title: record[2],
				Speakers: record[3], URL: record[4]}
			reports = append(reports, report)

		}

		isUpdated, isDeleted, errM := c.Database.InsertMany(c.Database.Collection("report"), reports)

		if errM != nil {
			return errM
		}

		if isUpdated {
			var wg sync.WaitGroup
			users, errS := c.Database.SelectUsers(c.Database.Collection("user"))
			if errS != nil {
				return errS
			}
			for _, user := range users {
				if user.TgID != int(ctx.EffectiveUser.Id) {
					msg, errSM := bot.SendMessage(ctx.EffectiveChat.Id, "Доклады были обновлены. Пожалуйста, ознакомьтесь с изменениями в \"👀 Посмотреть доклады\"", nil)
					if errSM != nil {
						return errM
					}
					wg.Add(1)
					go func(msg *gotgbot.Message) {
						defer wg.Done()
						time.Sleep(time.Second * 3)
						_, errD := bot.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
						if errD != nil {
							return
						}
					}(msg)
				}
			}
			wg.Wait()
		}

		if isDeleted {
			var wg sync.WaitGroup
			users, errS := c.Database.SelectUsers(c.Database.Collection("user"))
			if errS != nil {
				return errS
			}
			for _, user := range users {
				if user.TgID != int(ctx.EffectiveUser.Id) {
					msg, errSM := bot.SendMessage(ctx.EffectiveChat.Id, "Некоторые доклады были удалены. Пожалуйста, ознакомьтесь с новым списком докладов в \"👀 Посмотреть доклады\"", nil)
					if errSM != nil {
						return errSM
					}
					wg.Add(1)
					go func(msg *gotgbot.Message) {
						defer wg.Done()
						time.Sleep(time.Second * 3)
						_, errD := bot.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
						if errD != nil {
							return
						}
					}(msg)
				}
			}
			wg.Wait()
		}

		_, errS := bot.SendMessage(ctx.EffectiveChat.Id, "Ваше расписание успешно загружено!", &gotgbot.SendMessageOpts{
			ParseMode:   html,
			ReplyMarkup: backToMainMenuAdminKB(),
		})

		if errS != nil {
			return errS
		}
	default:
		_, errD := bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)

		if errD != nil {
			return errD
		}
	}
	return nil

}

func (c *Client) changeIdentificationCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	var err error

	err = c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, changeIdentification)

	if err != nil {
		return err
	}

	user, err := c.Database.SelectUser(c.Database.Collection("user"), int(ctx.EffectiveUser.Id))

	if err != nil {
		return err
	}

	cb := ctx.Update.CallbackQuery

	if _, _, err = cb.Message.EditText(bot, fmt.Sprintf("%s, введите, пожалуйста, ваш  билет/почту/ФИО (одно на выбор)", user.Identification),
		&gotgbot.EditMessageTextOpts{
			ParseMode:   html,
			ReplyMarkup: backToMainMenuKB(),
		}); err != nil {
		return err
	}

	return nil
}

func (c *Client) indexHandlerCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	if _, err := cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Это номер доклада"}); err != nil {
		return err
	}

	return nil
}

func (c *Client) viewReportsCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	var err error

	err = c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, viewReports)

	if err != nil {
		return err
	}

	data, err := c.Database.SelectReports(c.Database.Collection("report"))

	if err != nil {
		return err
	}

	cb := ctx.Update.CallbackQuery

	reportsFormat := getFormatReports(data)

	reports, errS := c.Database.SelectReports(c.Database.Collection("report"))

	if errS != nil {

		return errS
	}

	user, err := c.Database.SelectUser(c.Database.Collection("user"), int(ctx.EffectiveUser.Id))

	if err != nil {
		return err
	}

	_, _, err = cb.Message.EditText(bot, fmt.Sprintf("Доступные доклады:\n\n%s", reportsFormat), &gotgbot.EditMessageTextOpts{
		ReplyMarkup: reportsWithFavoriteKB(reports, user),
	})

	if err != nil {
		return err
	}

	return nil

}

func (c *Client) addToFavoriteCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	reports, errS := c.Database.SelectReports(c.Database.Collection("report"))

	if errS != nil {
		return errS
	}

	for _, report := range reports {
		if report.URL == strings.Split(cb.Data, ";")[1] {
			err := c.Database.AddUserFavReports(c.Database.Collection("user"), int(cb.From.Id), report)
			if err != nil {
				return err
			}
		}
	}

	user, errS := c.Database.SelectUser(c.Database.Collection("user"), int(cb.From.Id))

	if errS != nil {
		return errS
	}

	if _, _, err := cb.Message.EditReplyMarkup(bot, &gotgbot.EditMessageReplyMarkupOpts{ReplyMarkup: reportsWithFavoriteKB(reports, user)}); err != nil {
		return err
	}

	if _, err := cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Доклад успешно добавлен в избранное!"}); err != nil {
		return err
	}

	return nil
}

func (c *Client) removeFromFavoriteCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	reports, errS := c.Database.SelectReports(c.Database.Collection("report"))

	if errS != nil {

		return errS
	}

	for _, report := range reports {
		if report.URL == strings.Split(cb.Data, ";")[1] {
			err := c.Database.RemoveUserFavReport(c.Database.Collection("user"), int(cb.From.Id), report.URL)
			if err != nil {
				return err
			}
		}
	}

	user, errS := c.Database.SelectUser(c.Database.Collection("user"), int(cb.From.Id))

	if errS != nil {
		return errS
	}

	if _, _, err := cb.Message.EditReplyMarkup(bot, &gotgbot.EditMessageReplyMarkupOpts{ReplyMarkup: reportsWithFavoriteKB(reports, user)}); err != nil {
		return err
	}

	if _, err := cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Доклад убран из избранного!"}); err != nil {
		return err
	}

	return nil
}

func (c *Client) threePointsCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	if _, err := cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Доступных докладов пока нет"}); err != nil {

		return err
	}

	return nil
}

func (c *Client) photoHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	_, err := bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)

	if err != nil {

		return err
	}

	return nil
}

func (c *Client) audioHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	_, err := bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)

	if err != nil {

		return err
	}

	return nil
}

func (c *Client) videoHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	_, err := bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)

	if err != nil {

		return err
	}

	return nil
}

func (c *Client) mediaGroupHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	_, err := bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)

	if err != nil {

		return err
	}

	return nil
}

func (c *Client) storyHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	_, err := bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)

	if err != nil {

		return err
	}

	return nil
}

func (c *Client) videoNoteHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	_, err := bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)

	if err != nil {

		return err
	}

	return nil
}

func (c *Client) notEvaluateCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	if _, err := cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Вы пока не можете оценить этот доклад"}); err != nil {

		return err
	}

	return nil
}

func (c *Client) evaluateReportCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	url := strings.Split(cb.Data, ";")[1]

	report, err := c.Database.SelectReport(c.Database.Collection("report"), url)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("Вы оцениваете следующий доклад:\n\n%s - %s\n\nКакую оценку вы бы поставили за содержание доклада:", report.Speakers, report.Title)
	_, _, errE := cb.Message.EditText(bot, text,
		&gotgbot.EditMessageTextOpts{
			ReplyMarkup: evaluateKB(),
		})

	if errE != nil {
		return errE
	}

	if err := c.FSM.SetState(context.Background(), cb.From.Id, fmt.Sprintf("evaluateReport;%s;%s", url, text)); err != nil {
		return err
	}

	return nil
}

func (c *Client) contentCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	state, err := c.FSM.GetState(context.Background(), cb.From.Id)

	if err != nil {
		return err
	}

	stateSeparated := strings.Split(state, ";")

	_, _, err = cb.Message.EditReplyMarkup(bot, &gotgbot.EditMessageReplyMarkupOpts{
		ReplyMarkup: contentKB(stateSeparated[1]),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) performanceCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	markForContent := strings.Split(cb.Data, ";")[1]

	_, _, err := cb.Message.EditText(bot, "Какую бы оценку вы поставили за выступление:", &gotgbot.EditMessageTextOpts{
		ParseMode:   html,
		ReplyMarkup: performanceKB(),
	})

	if err != nil {
		return err
	}

	state, err := c.FSM.GetState(context.Background(), cb.From.Id)

	if err != nil {
		return err
	}

	if err := c.FSM.SetState(context.Background(), cb.From.Id, fmt.Sprintf("%s;%s", state, markForContent)); err != nil {
		return err
	}

	return nil
}

func (c *Client) backToContentCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	state, err := c.FSM.GetState(context.Background(), cb.From.Id)

	if err != nil {

	}

	stateSeparated := strings.Split(state, ";")

	if stateSeparated[0] == evaluateReport {

		if err != nil {
			return err
		}

		_, _, errE := cb.Message.EditText(bot, stateSeparated[2], &gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveChat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			ReplyMarkup: evaluateKB(),
		})

		if errE != nil {
			return errE
		}

		err := c.FSM.SetState(context.Background(), cb.From.Id, strings.Join(stateSeparated[:3], ";")+";")

		if err != nil {
			return err
		}

	}

	return nil
}

func (c *Client) customMsgCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	markPerformance := strings.Split(cb.Data, ";")[1]

	state, err := c.FSM.GetState(context.Background(), cb.From.Id)

	if err != nil {
		return err
	}

	errS := c.FSM.SetState(context.Background(), cb.From.Id, fmt.Sprintf("%s;%s", state, markPerformance))

	if errS != nil {
		return errS
	}

	_, _, err = cb.Message.EditText(bot, "Введите дополнительный комментарий или нажмите на кнопку \"Далее\"", &gotgbot.EditMessageTextOpts{
		ReplyMarkup: commentKB(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) evaluateEndNoCommentCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	// todo adsa
	cb := ctx.Update.CallbackQuery

	state, err := c.FSM.GetState(context.Background(), cb.From.Id)

	if err != nil {
		return err
	}

	stateSeparated := strings.Split(state, ";")

	evaluation := models.Evaluation{URL: stateSeparated[1], TgID: int(cb.From.Id),
		Content: stateSeparated[len(stateSeparated)-2], Performance: stateSeparated[len(stateSeparated)-1]}
	if err = c.Database.InsertOne(c.Database.Collection("evaluation"), evaluation); err != nil {
		return err
	}
	_, _, err = cb.Message.EditText(bot, "Ваш отзыв успешно добавлен!", &gotgbot.EditMessageTextOpts{
		ReplyMarkup: afterMarkKB(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) noWishToMarkCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	state, err := c.FSM.GetState(context.Background(), cb.From.Id)

	if err != nil {
		return err
	}

	stateSeparated := strings.Split(state, ";")

	evaluation := models.Evaluation{URL: stateSeparated[1], TgID: int(cb.From.Id),
		Content: cb.Data}

	if err = c.Database.InsertOne(c.Database.Collection("evaluation"), evaluation); err != nil {
		return err
	}

	_, _, err = cb.Message.EditText(bot, "Спасибо за вашу обратную связь! Помните: вы всегда можете изменить свой отзыв", &gotgbot.EditMessageTextOpts{
		ReplyMarkup: afterMarkKB(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) noMarkCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	state, err := c.FSM.GetState(context.Background(), cb.From.Id)

	if err != nil {
		return err
	}

	stateSeparated := strings.Split(state, ";")

	evaluation := models.Evaluation{URL: stateSeparated[1], TgID: int(cb.From.Id),
		Content: cb.Data}

	if err = c.Database.InsertOne(c.Database.Collection("evaluation"), evaluation); err != nil {
		return err
	}

	_, _, err = cb.Message.EditText(bot, "Спасибо за ваш отзыв, вдруг что, вы всегда можете его изменить", &gotgbot.EditMessageTextOpts{
		ReplyMarkup: afterMarkKB(),
	})

	if err != nil {
		return err
	}

	return nil
}
