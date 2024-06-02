package handlers

import (
	"fmt"
	"github.com/NOSTRADA88/telegram-bot-go/internal/bot/fsm"
	"github.com/NOSTRADA88/telegram-bot-go/internal/config"
	"github.com/NOSTRADA88/telegram-bot-go/internal/repository/mongodb"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

const (
	start                = "start"
	menu                 = "menu"
	confInfo             = "confInfo"
	viewReports          = "viewReports"
	updateIdentification = "updateIdentification"
	uploadSchedule       = "uploadSchedule"
	back                 = "back"
	downloadReviews      = "downloadReviews"
	index                = "index"
	evaluateReport       = "evaluateReport"
	notEvaluateReport    = "notEvaluateReport"
	evaluationBegin      = "evaluationBegin"
	noWishToEvaluate     = "noWishToEvaluate"
	noEvaluate           = "noEvaluate"
	evaluationEnd        = "evaluationEnd"
	backToContent        = "backToContent"
	threePoints          = "..."
	userEvaluations      = "userEvaluations"
	updateEvaluation     = "updateEvaluation"
	deleteEvaluation     = "deleteEvaluation"
	updateContent        = "updateContent"
	updatePerformance    = "updatePerformance"
	updateComment        = "updateComment"
	updateNoComment      = "updateNoComment"
)

func Set(dispatcher *ext.Dispatcher, c Client) {
	dispatcher.AddHandler(handlers.NewCommand(start, c.startHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(confInfo), c.confInfoCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(viewReports), c.viewReportsCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(updateIdentification), c.changeIdentificationCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(uploadSchedule), c.uploadScheduleCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(back), c.backCBHandler))
	dispatcher.AddHandler(handlers.NewMessage(message.Text, c.textHandler))
	dispatcher.AddHandler(handlers.NewMessage(message.Document, c.fileHandler))
	dispatcher.AddHandler(handlers.NewMessage(message.Photo, c.photoHandler))
	dispatcher.AddHandler(handlers.NewMessage(message.Audio, c.audioHandler))
	dispatcher.AddHandler(handlers.NewMessage(message.Video, c.videoHandler))
	dispatcher.AddHandler(handlers.NewMessage(message.MediaGroup, c.mediaGroupHandler))
	dispatcher.AddHandler(handlers.NewMessage(message.Sticker, c.mediaGroupHandler))
	dispatcher.AddHandler(handlers.NewMessage(message.Story, c.storyHandler))
	dispatcher.AddHandler(handlers.NewMessage(message.VideoNote, c.videoNoteHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(index), c.indexHandlerCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("add;"), c.addToFavoriteCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("remove;"), c.removeFromFavoriteCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(threePoints), c.threePointsCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(notEvaluateReport), c.notEvaluateCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(evaluateReport), c.evaluateReportCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(evaluationBegin), c.evaluationBeginCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("content;"), c.contentCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(backToContent), c.backToContentCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("performance;"), c.performanceCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(evaluationEnd), c.evaluateEndNoCommentCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(noWishToEvaluate), c.noWishToEvaluateCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(noEvaluate), c.noEvaluateCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(userEvaluations), c.userEvaluationsCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(fmt.Sprintf("%s;", updateEvaluation)), c.updateEvaluationCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(fmt.Sprintf("%s;", deleteEvaluation)), c.deleteEvaluationCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(fmt.Sprintf("%s;", updateContent)), c.updateContentCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(fmt.Sprintf("%s;", updatePerformance)), c.updatePerformanceCBHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(updateNoComment), c.updateWithNoCommentCBHandler))
}

type Client struct {
	Cfg      *config.Config
	FSM      fsm.StateController
	Database mongodb.DataManipulator
}
