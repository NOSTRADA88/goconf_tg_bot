package handlers

import (
	"fmt"
	"github.com/NOSTRADA88/telegram-bot-go/internal/models"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"strconv"
	"time"
)

func mainMenuAdminKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "📋 Информация о конференции", CallbackData: confInfo},
		},
		{
			{Text: "👀 Посмотреть доклады", CallbackData: viewReports},
		},
		{
			{Text: "📝 Редактировать идентификацию", CallbackData: updateIdentification},
		},
		{
			{Text: "📥 Загрузить расписание", CallbackData: uploadSchedule},
		},
		{
			{Text: "📂 Выгрузить файл с оценками", CallbackData: downloadReviews},
		},
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func mainMenuUserKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "📋 Информация о конференции", CallbackData: confInfo},
		},
		{
			{Text: "👀 Посмотреть доклады", CallbackData: viewReports},
		},
		{
			{Text: "📝 Редактировать идентификацию", CallbackData: updateIdentification},
		},
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func backToMainMenuKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "⬅️ Назад", CallbackData: back},
		},
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func backToMainMenuAdminKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "⬅️ Венуться на главную", CallbackData: back},
		},
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func reportsWithFavoriteKB(reports []models.Report, user models.User, evaluations []models.Evaluation) gotgbot.InlineKeyboardMarkup {

	if len(reports) == 0 {
		kb := [][]gotgbot.InlineKeyboardButton{
			{
				{Text: threePoints, CallbackData: threePoints},
			},
			{
				{Text: "⬅️ Назад", CallbackData: back},
			},
		}
		return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
	}

	var kb [][]gotgbot.InlineKeyboardButton

	location, err := time.LoadLocation("Europe/Moscow")

	if err != nil {
		println(err)
	}

	if len(user.FavoriteReports) == 0 {
		for ind, report := range reports {

			startTime := report.StartTime.Truncate(time.Second)

			now := time.Now().In(location).Truncate(time.Second)

			reportMSKTime := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(),
				startTime.Minute(), startTime.Second(), startTime.Nanosecond(), location)

			evl := "⛔"
			evlCB := notEvaluateReport

			if reportMSKTime.Before(now) || startTime.Equal(now) {
				evl = "🏆"
				evlCB = fmt.Sprintf("%s;%s", evaluateReport, report.URL)
			}

			kb = append(kb, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("%v.", ind+1), CallbackData: "index"},
				{Text: "⏳", CallbackData: "nothing", Url: report.URL},
				{Text: fmt.Sprintf("%v м", strconv.Itoa(report.Duration)), Url: report.URL, CallbackData: "nothing"},
				{Text: "⭐", CallbackData: fmt.Sprintf("add;%s", report.URL)},
				{Text: evl, CallbackData: evlCB},
			})

		}

		if len(evaluations) != 0 {
			kb = append(kb, []gotgbot.InlineKeyboardButton{
				{Text: "Мои отзывы", CallbackData: userEvaluations},
			})
		}

	} else {
		favReports := make(map[string]bool, len(reports))

		for _, report := range user.FavoriteReports {

			favReports[report.URL] = true

		}

		for ind, report := range reports {

			startTime := report.StartTime.Truncate(time.Second)

			now := time.Now().In(location).Truncate(time.Second)

			if err != nil {
				fmt.Println(err)
			}

			reportMSKTime := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(),
				startTime.Minute(), startTime.Second(), startTime.Nanosecond(), location)

			_, isFav := favReports[report.URL]

			favText := "⭐"

			cb := fmt.Sprintf("add;%s", report.URL)

			if isFav {
				favText = "🌟"
				cb = fmt.Sprintf("remove;%s", report.URL)
			}

			evl := "⛔"
			evlCB := notEvaluateReport

			if reportMSKTime.Before(now) || startTime.Equal(now) {
				evl = "🏆"
				evlCB = fmt.Sprintf("%s;%s", evaluateReport, report.URL)
			}

			kb = append(kb, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("%v.", ind+1), CallbackData: "index"},
				{Text: "⏳", CallbackData: "nothing", Url: report.URL},
				{Text: fmt.Sprintf("%v м", strconv.Itoa(report.Duration)), Url: report.URL, CallbackData: "nothing"},
				{Text: favText, CallbackData: cb},
				{Text: evl, CallbackData: evlCB},
			})
		}

		if len(evaluations) != 0 {
			kb = append(kb, []gotgbot.InlineKeyboardButton{
				{Text: "Мои отзывы", CallbackData: userEvaluations},
			})
		}

	}

	kb = append(kb, []gotgbot.InlineKeyboardButton{
		{Text: "⬅️ Назад", CallbackData: back},
	})

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func evaluateKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "Оценить от 1 до 5", CallbackData: evaluationBegin},
		},
		{
			{Text: "Я не слушал этот доклад", CallbackData: noEvaluate},
		},
		{
			{Text: "Я не хочу оценивать этот доклад", CallbackData: noWishToEvaluate},
		},
		{
			{Text: "Вернуться к докладам", CallbackData: viewReports},
		},
	}

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func contentKB(url string) gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "1", CallbackData: "content;1"}, {Text: "2", CallbackData: "content;2"},
			{Text: "3", CallbackData: "content;3"}, {Text: "4", CallbackData: "content;4"}, {Text: "5", CallbackData: "content;5"},
		},
		{
			{Text: "Назад", CallbackData: fmt.Sprintf("%s;%s", evaluateReport, url)},
		},
	}

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func performanceKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "1", CallbackData: "performance;1"}, {Text: "2", CallbackData: "performance;2"},
			{Text: "3", CallbackData: "performance;3"}, {Text: "4", CallbackData: "performance;4"}, {Text: "5", CallbackData: "performance;5"},
		},
		{
			{Text: "Вернуться в к выбору оценки", CallbackData: "backToContent"},
		},
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func commentKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "Далее", CallbackData: evaluationEnd},
		},
		{
			{Text: "Вернуться к выбору оценки", CallbackData: backToContent},
		},
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func evaluationEndKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "К докладам", CallbackData: viewReports},
		},
		{
			{Text: "В главное меню", CallbackData: back},
		},
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func userEvaluationsKB(reports []models.Report, evaluationMap map[string]models.Evaluation) gotgbot.InlineKeyboardMarkup {
	var kb [][]gotgbot.InlineKeyboardButton

	for ind, report := range reports {
		if _, exists := evaluationMap[report.URL]; exists {
			updCB := fmt.Sprintf("%s;%s", updateEvaluation, report.URL)
			dltCB := fmt.Sprintf("%s;%s", deleteEvaluation, report.URL)
			kb = append(kb, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("%v.", ind+1), CallbackData: "index"},
				{Text: "✏️ Редактировать", CallbackData: updCB},
				{Text: "🗑️ Удалить", CallbackData: dltCB},
			})
		}
	}

	kb = append(kb, []gotgbot.InlineKeyboardButton{
		{Text: "Вернуться к докладам", CallbackData: viewReports},
	})

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func contentUpdateKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "1", CallbackData: "updateContent;1"}, {Text: "2", CallbackData: "updateContent;2"},
			{Text: "3", CallbackData: "updateContent;3"}, {Text: "4", CallbackData: "updateContent;4"}, {Text: "5", CallbackData: "updateContent;5"},
		},
		{
			{Text: "К отзывам", CallbackData: userEvaluations},
		},
	}

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func performanceUpdateKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "1", CallbackData: "updatePerformance;1"}, {Text: "2", CallbackData: "updatePerformance;2"},
			{Text: "3", CallbackData: "updatePerformance;3"}, {Text: "4", CallbackData: "updatePerformance;4"}, {Text: "5", CallbackData: "updatePerformance;5"},
		},
		{
			{Text: "К отзывам", CallbackData: userEvaluations},
		},
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func commentUpdateKB() gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "Далее", CallbackData: updateNoComment},
		},
		{
			{Text: "К отзывам", CallbackData: userEvaluations},
		},
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}
