package handlers

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/NOSTRADA88/telegram-bot-go/internal/models"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	html     = "html"
	dataJSON = "data.json"
)

func (c *Client) startHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

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

	case updateIdentification:

		err = c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, menu)

		if err != nil {
			return err
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

			_, err = bot.SendMessage(ctx.EffectiveChat.Id, "Простите, но имя не может начинаться с \"/\". Введите, что-нибудь другое... билет... ФИО или вашу почту.", nil)

			if err != nil {
				return err
			}

			return nil
		}
	default:
		err = c.FSM.SetState(context.Background(), ctx.Message.From.Id, menu)

		if err != nil {
			return err
		}

		user, errS := c.Database.SelectUser(c.Database.Collection("user"), int(ctx.EffectiveUser.Id))

		if errS != nil {
			return errS
		}

		if _, exists := c.Cfg.Administrators.IDsInMap[int(ctx.Message.From.Id)]; exists {
			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				fmt.Sprintf("Приветствую %s. Вы уже успели ознакомится со списком докладов ? Если нет, то крайне рекомендую, сегодня выступают отличные спикеры!", user.Identification),
				&gotgbot.SendMessageOpts{
					ParseMode:   html,
					ReplyMarkup: mainMenuAdminKB(),
				})

			if err != nil {
				return err
			}

		} else {
			_, err = bot.SendMessage(ctx.Message.Chat.Id,
				fmt.Sprintf("Приветствую %s. Вы уже успели ознакомится со списком докладов ? Если нет, то крайне рекомендую, сегодня выступают отличные спикеры!", user.Identification),
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

			_, errS := bot.SendMessage(ctx.EffectiveChat.Id, "Простите, но имя не может начинаться с \"/\". Введите, билет, ФИО или вашу почту.",
				nil)

			if errS != nil {
				return errS
			}

			return nil
		}

		coll := c.Database.Collection("user")

		err = c.Database.InsertOne(coll, models.User{
			TgID: int(ctx.EffectiveUser.Id), Identification: ctx.EffectiveMessage.Text, FavoriteReports: []models.Report{}, ChatID: int(ctx.EffectiveChat.Id)})

		if err != nil {
			return err
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
	case updateIdentification:

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
	case uploadSchedule, viewReports, userEvaluations:
		_, errD := bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)

		if errD != nil {
			return errD
		}
	default:
		if len(strings.Split(state, ";")) == 5 && strings.Split(state, ";")[0] == evaluateReport {

			stateSeparated := strings.Split(state, ";")
			text := ctx.EffectiveMessage.Text
			evaluation := models.Evaluation{URL: stateSeparated[1], TgID: int(ctx.Message.From.Id),
				Content: stateSeparated[len(stateSeparated)-2], Performance: stateSeparated[len(stateSeparated)-1],
				Comment: text}

			if err = c.Database.InsertOne(c.Database.Collection("evaluation"), evaluation); err != nil {
				return err
			}

			_, err = bot.SendMessage(ctx.EffectiveChat.Id, "Ваш отзыв успешно добавлен!", &gotgbot.SendMessageOpts{
				ReplyMarkup: evaluationEndKB(),
				ParseMode:   html,
			})

			if err != nil {
				return err
			}
		}

		if len(strings.Split(state, ";")) == 4 && strings.Split(state, ";")[0] == updateEvaluation {

			stateSeparated := strings.Split(state, ";")

			text := ctx.EffectiveMessage.Text

			evaluation := models.Evaluation{URL: stateSeparated[1], TgID: int(ctx.Message.From.Id),
				Content: stateSeparated[2], Performance: stateSeparated[3],
				Comment: text}

			upd, err := c.Database.UpdateEvaluation(c.Database.Collection("evaluation"), int(ctx.Message.From.Id), stateSeparated[1], evaluation)

			if err != nil {
				return err
			}

			err = c.FSM.SetState(context.Background(), ctx.Message.From.Id, updateComment)

			if err != nil {
				return err
			}

			if upd {
				_, err = bot.SendMessage(ctx.EffectiveChat.Id, "Ваш отзыв успешно обновлён!", &gotgbot.SendMessageOpts{
					ReplyMarkup: evaluationEndKB(),
					ParseMode:   html,
				})
				if err != nil {
					return err
				}
			} else {
				_, err = bot.SendMessage(ctx.EffectiveChat.Id, "Ваш отзыв ни чем не отличается от прошлого!", &gotgbot.SendMessageOpts{
					ReplyMarkup: evaluationEndKB(),
					ParseMode:   html,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *Client) confInfoCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	err := c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, confInfo)
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

	err := c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, menu)

	if err != nil {
		return err
	}

	cb := ctx.Update.CallbackQuery

	if _, exists := c.Cfg.Administrators.IDsInMap[int(cb.From.Id)]; exists {
		_, _, err = cb.Message.EditText(bot,
			"Вы вернулись в главное меню. Как удобно, что я обрабатываю все ваши сценрии использования.",
			&gotgbot.EditMessageTextOpts{ParseMode: html, ReplyMarkup: mainMenuAdminKB()})

		if err != nil {
			return err
		}

	} else {
		_, _, err = cb.Message.EditText(bot,
			"Вы вернулись в главное меню. Как удобно, что я обрабатываю все ваши сценрии использования.",
			&gotgbot.EditMessageTextOpts{ParseMode: html, ReplyMarkup: mainMenuUserKB()})

		if err != nil {
			return err
		}

	}

	return nil
}

func (c *Client) uploadScheduleCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	err := c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, uploadSchedule)

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
					msg, errSM := bot.SendMessage(ctx.EffectiveChat.Id, "Некоторые доклады были удалены. Пожалуйста, ознакомьтесь с изменённым списком докладов в \"👀 Посмотреть доклады\"", nil)
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

		_, err = bot.SendMessage(ctx.EffectiveChat.Id, "Ваше расписание успешно загружено!", &gotgbot.SendMessageOpts{
			ParseMode:   html,
			ReplyMarkup: backToMainMenuAdminKB(),
		})

		if err != nil {
			return err
		}
	default:
		_, err = bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)

		if err != nil {
			return err
		}
	}
	return nil

}

func (c *Client) changeIdentificationCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	err := c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, updateIdentification)

	if err != nil {
		return err
	}

	user, err := c.Database.SelectUser(c.Database.Collection("user"), int(ctx.EffectiveUser.Id))

	if err != nil {
		return err
	}

	cb := ctx.Update.CallbackQuery

	if _, _, err = cb.Message.EditText(bot, fmt.Sprintf("Сейчас вы известным мне как %s. Введите, пожалуйста, ваш  билет/почту/ФИО (одно на выбор)", user.Identification),
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

	if _, err := cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Это номер доклада в программе"}); err != nil {
		return err
	}

	return nil
}

func (c *Client) viewReportsCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	err := c.FSM.SetState(context.Background(), ctx.EffectiveUser.Id, viewReports)

	if err != nil {
		return err
	}

	data, err := c.Database.SelectReports(c.Database.Collection("report"))

	if err != nil {
		return err
	}

	cb := ctx.Update.CallbackQuery

	reportsFormat := getFormatReports(data)

	reports, err := c.Database.SelectReports(c.Database.Collection("report"))

	if err != nil {

		return err
	}

	user, err := c.Database.SelectUser(c.Database.Collection("user"), int(ctx.EffectiveUser.Id))

	if err != nil {
		return err
	}

	evaluations, err := c.Database.SelectEvaluations(c.Database.Collection("evaluation"), int(cb.From.Id))

	if err != nil {
		return err
	}

	_, _, err = cb.Message.EditText(bot, fmt.Sprintf("Доступные доклады:\n\n%s", reportsFormat), &gotgbot.EditMessageTextOpts{
		ReplyMarkup: reportsWithFavoriteKB(reports, user, evaluations),
	})

	if err != nil {
		return err
	}

	return nil

}

func (c *Client) addToFavoriteCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	reports, err := c.Database.SelectReports(c.Database.Collection("report"))

	if err != nil {
		return err
	}

	for _, report := range reports {
		if report.URL == strings.Split(cb.Data, ";")[1] {
			err = c.Database.AddUserFavReports(c.Database.Collection("user"), int(cb.From.Id), report)
			if err != nil {
				return err
			}
		}
	}

	user, err := c.Database.SelectUser(c.Database.Collection("user"), int(cb.From.Id))

	if err != nil {
		return err
	}

	evaluations, err := c.Database.SelectEvaluations(c.Database.Collection("evaluation"), int(cb.From.Id))

	if err != nil {
		return err
	}

	if _, _, err = cb.Message.EditReplyMarkup(bot, &gotgbot.EditMessageReplyMarkupOpts{ReplyMarkup: reportsWithFavoriteKB(reports, user, evaluations)}); err != nil {
		return err
	}

	if _, err = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Доклад успешно добавлен в избранное!"}); err != nil {
		return err
	}

	return nil
}

func (c *Client) removeFromFavoriteCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	reports, err := c.Database.SelectReports(c.Database.Collection("report"))

	if err != nil {

		return err
	}

	for _, report := range reports {
		if report.URL == strings.Split(cb.Data, ";")[1] {
			err = c.Database.RemoveUserFavReport(c.Database.Collection("user"), int(cb.From.Id), report.URL)
			if err != nil {
				return err
			}
		}
	}

	user, err := c.Database.SelectUser(c.Database.Collection("user"), int(cb.From.Id))

	if err != nil {
		return err
	}

	evaluations, err := c.Database.SelectEvaluations(c.Database.Collection("evaluation"), int(cb.From.Id))

	if err != nil {
		return err
	}

	if _, _, err = cb.Message.EditReplyMarkup(bot, &gotgbot.EditMessageReplyMarkupOpts{ReplyMarkup: reportsWithFavoriteKB(reports, user, evaluations)}); err != nil {
		return err
	}

	if _, err = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Доклад убран из избранного!"}); err != nil {
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
	_, _, err = cb.Message.EditText(bot, text,
		&gotgbot.EditMessageTextOpts{
			ReplyMarkup: evaluateKB(),
		})

	if err != nil {
		return err
	}

	if err = c.FSM.SetState(context.Background(), cb.From.Id, fmt.Sprintf("evaluateReport;%s;%s", url, text)); err != nil {
		return err
	}

	return nil
}

func (c *Client) evaluationBeginCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

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

func (c *Client) contentCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

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

	if err = c.FSM.SetState(context.Background(), cb.From.Id, fmt.Sprintf("%s;%s", state, markForContent)); err != nil {
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

		_, _, err = cb.Message.EditText(bot, stateSeparated[2], &gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveChat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			ReplyMarkup: evaluateKB(),
		})

		if err != nil {
			return err
		}

		err = c.FSM.SetState(context.Background(), cb.From.Id, strings.Join(stateSeparated[:3], ";")+";")

		if err != nil {
			return err
		}

	}

	return nil
}

func (c *Client) performanceCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

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
		ReplyMarkup: evaluationEndKB(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) noWishToEvaluateCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

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
		ReplyMarkup: evaluationEndKB(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) noEvaluateCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

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
		ReplyMarkup: evaluationEndKB(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) userEvaluationsCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	err := c.FSM.SetState(context.Background(), cb.From.Id, userEvaluations)

	if err != nil {
		return err
	}

	evaluations, err := c.Database.SelectEvaluations(c.Database.Collection("evaluation"), int(cb.From.Id))

	var text string

	reports, err := c.Database.SelectReports(c.Database.Collection("report"))

	evaluationsMap := make(map[string]models.Evaluation, len(evaluations))

	for _, evaluation := range evaluations {
		if _, exists := evaluationsMap[evaluation.URL]; !exists {
			evaluationsMap[evaluation.URL] = evaluation
		}
	}

	if err != nil {
		return err
	}

	for ind, report := range reports {
		if _, exists := evaluationsMap[report.URL]; exists {
			text += fmt.Sprintf("%v. %s - %s\n\nСодержание: \"%s\"\nВыступление: \"%s\"\nКомментарий: \"%s\"\n\n", ind+1,
				report.Speakers, report.Title, evaluationsMap[report.URL].Content, evaluationsMap[report.URL].Performance, evaluationsMap[report.URL].Comment)
		}
	}

	_, _, err = cb.Message.EditText(bot, fmt.Sprintf("Ваши отзывы:\n\n%s", text), &gotgbot.EditMessageTextOpts{
		ReplyMarkup: userEvaluationsKB(reports, evaluationsMap),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) updateEvaluationCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	cbSeparated := strings.Split(cb.Data, ";")

	url := cbSeparated[1]

	report, err := c.Database.SelectReport(c.Database.Collection("report"), url)

	if err != nil {
		return err
	}

	err = c.FSM.SetState(context.Background(), cb.From.Id, fmt.Sprintf("%s;%s", cbSeparated[0], url))

	if err != nil {
		return err
	}

	_, _, err = cb.Message.EditText(bot, fmt.Sprintf("Выберите оценку:\n\n%s - %s", report.Speakers, report.Title), &gotgbot.EditMessageTextOpts{
		ReplyMarkup: contentUpdateKB(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) deleteEvaluationCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	cbSeparated := strings.Split(cb.Data, ";")

	err := c.FSM.SetState(context.Background(), cb.From.Id, cbSeparated[0])

	if err != nil {
		return err
	}

	url := cbSeparated[1]

	deleted, err := c.Database.DeleteEvaluation(c.Database.Collection("evaluation"), int(cb.From.Id), url)

	if err != nil {
		return err
	}

	if deleted {
		_, _, err = cb.Message.EditText(bot, "Ваш отзыв удалён! Если передумайте, то всегда можете написать новый", &gotgbot.EditMessageTextOpts{
			ReplyMarkup: evaluationEndKB(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) updateContentCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	cbSeparated := strings.Split(cb.Data, ";")

	content := cbSeparated[1]

	state, err := c.FSM.GetState(context.Background(), cb.From.Id)

	if err != nil {
		return err
	}

	err = c.FSM.SetState(context.Background(), cb.From.Id, fmt.Sprintf("%s;%s", state, content))

	_, _, err = cb.Message.EditText(bot, "Введи вашу оценку за выступление: ", &gotgbot.EditMessageTextOpts{
		ReplyMarkup: performanceUpdateKB(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) updatePerformanceCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	cbSeparated := strings.Split(cb.Data, ";")

	content := cbSeparated[1]

	state, err := c.FSM.GetState(context.Background(), cb.From.Id)

	if err != nil {
		return err
	}

	err = c.FSM.SetState(context.Background(), cb.From.Id, fmt.Sprintf("%s;%s", state, content))

	_, _, err = cb.Message.EditText(bot, "Введите дополнительный комментарий или нажмите на кнопку \"Далее\"", &gotgbot.EditMessageTextOpts{
		ReplyMarkup: commentUpdateKB(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) updateWithNoCommentCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	state, err := c.FSM.GetState(context.Background(), cb.From.Id)

	if err != nil {
		return err
	}

	stateSeparated := strings.Split(state, ";")

	evaluation := models.Evaluation{
		Content: stateSeparated[2], Performance: stateSeparated[3],
	}

	upd, err := c.Database.UpdateEvaluation(c.Database.Collection("evaluation"), int(cb.From.Id), stateSeparated[1], evaluation)

	if upd {
		_, _, err = cb.Message.EditText(bot, "Ваш отзыв успешно обновлён!", &gotgbot.EditMessageTextOpts{
			ReplyMarkup: evaluationEndKB(),
		})

		if err != nil {
			return err
		}
	} else {
		_, _, err = cb.Message.EditText(bot, "Обновление не случилось. Скорее всего вы ввели такие же оценки", &gotgbot.EditMessageTextOpts{
			ReplyMarkup: evaluationEndKB(),
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) checkAndNotify(bot *gotgbot.Bot) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := c.notifyUpcomingReports(bot)
			if err != nil {
				fmt.Println("Error notifying users:", err)
			}
		}
	}
}

func (c *Client) notifyUpcomingReports(bot *gotgbot.Bot) error {
	reports, err := c.Database.SelectReports(c.Database.Collection("report"))
	if err != nil {
		return err
	}

	users, err := c.Database.SelectUsers(c.Database.Collection("user"))
	if err != nil {
		return err
	}

	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return err
	}

	now := time.Now().In(location)
	for _, report := range reports {
		startTime := report.StartTime.In(location)
		if startTime.After(now) && startTime.Before(now.Add(10*time.Minute)) {
			message := fmt.Sprintf("Доклад \"%s\" начнется через 10 минут.\nСпикер: %s\nНачало в %s", report.Title, report.Speakers, startTime.Format("15:04"))
			for _, user := range users {
				_, err := bot.SendMessage(int64(user.TgID), message, nil)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *Client) downloadReviewsCBHandler(bot *gotgbot.Bot, ctx *ext.Context) error {

	cb := ctx.Update.CallbackQuery

	reports, err := c.Database.SelectReports(c.Database.Collection("report"))

	if err != nil {
		return err
	}

	evaluations, err := c.Database.SelectAllEvaluations(c.Database.Collection("evaluation"))

	if err != nil {
		return err
	}

	evaluationsMap := make(map[string]models.Evaluation, len(evaluations))

	for _, evaluation := range evaluations {
		if _, exists := evaluationsMap[evaluation.URL]; !exists {
			evaluationsMap[evaluation.URL] = evaluation
		}
	}

	var actualEvaluations []models.Evaluation

	for _, report := range reports {
		if evaluation, exists := evaluationsMap[report.URL]; exists {
			actualEvaluations = append(actualEvaluations, evaluation)
		}
	}

	jsonData, err := json.Marshal(actualEvaluations)

	if err != nil {
		return err
	}

	file, err := os.Create(dataJSON)

	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)

	if err != nil {
		return err
	}

	err = file.Close()

	if err != nil {
		return err
	}

	file, err = os.Open(dataJSON)

	if err != nil {
		return err
	}

	defer func() {
		err = os.Remove(file.Name())
		if err != nil {
			fmt.Println(err)
		}
	}()
	var wg sync.WaitGroup
	msg, err := bot.SendDocument(cb.From.Id, file, &gotgbot.SendDocumentOpts{Caption: "Отзывы для текущих докладов"})

	if err != nil {
		return err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(60 * time.Second)
		_, err = bot.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	wg.Wait()
	return nil
}
