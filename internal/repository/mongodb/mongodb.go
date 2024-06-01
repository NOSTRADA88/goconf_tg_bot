package mongodb

import (
	"context"
	"fmt"
	"github.com/NOSTRADA88/telegram-bot-go/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	mongo *mongo.Client
}

type DataManipulator interface {
	Init() error
	Collection(collection string) *mongo.Collection
	InsertOne(coll *mongo.Collection, data interface{}) error
	InsertMany(coll *mongo.Collection, data []interface{}) (bool, bool, error)
	SelectReport(coll *mongo.Collection, url string) (models.Report, error)
	SelectReports(coll *mongo.Collection) ([]models.Report, error)
	SelectUser(coll *mongo.Collection, tgID int) (models.User, error)
	SelectUsers(coll *mongo.Collection) ([]models.User, error)
	UpdateUserID(coll *mongo.Collection, tgID int, identification string) (bool, error)
	AddUserFavReports(coll *mongo.Collection, tgID int, report models.Report) error
	RemoveUserFavReport(coll *mongo.Collection, tgID int, reportURL string) error
	SelectEvaluation(coll *mongo.Collection, tgID int, url string) (models.Evaluation, error)
	SelectEvaluations(coll *mongo.Collection, tgID int) ([]models.Evaluation, error)
	SelectAllEvaluations(coll *mongo.Collection) ([]models.Evaluation, error)
	UpdateEvaluation(coll *mongo.Collection, tgID int, url string, evaluation models.Evaluation) (bool, error)
}

func New(host string, port int, user, password string) (*Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%v", user, password, host, port)))
	if err != nil {
		return nil, err
	}
	return &Client{mongo: client}, nil
}

func (c *Client) Close() error {
	return c.mongo.Disconnect(context.Background())
}

func (c *Client) Init() error {
	if err := c.ensureUserTgIDUnique(); err != nil {
		return err
	}
	return c.ensureEvaluationTgIDAndURLUnique()
}

func (c *Client) ensureUserTgIDUnique() error {
	coll := c.Collection("user")
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"tgID": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err := coll.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return err
	}
	return err
}

func (c *Client) ensureEvaluationTgIDAndURLUnique() error {
	coll := c.Collection("evaluation")
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "tgID", Value: 1},
			{Key: "url", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := coll.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return err
	}
	return err
}

func (c *Client) Collection(collection string) *mongo.Collection {
	return c.mongo.Database("tg-bot").Collection(collection)
}

func (c *Client) InsertMany(coll *mongo.Collection, data []interface{}) (bool, bool, error) {
	switch coll.Name() {
	case "report":
		existingURLS := make([]string, len(data), len(data))
		for ind, report := range data {
			existingURLS[ind] = report.(models.Report).URL
		}
		isUpdated, err := updateReports(coll, data)
		if err != nil {
			return false, false, err
		}
		isDeleted, err := deleteReports(coll, existingURLS)
		if err != nil {
			return false, false, err
		}
		return isUpdated, isDeleted, nil
	default:
		_, err := coll.InsertMany(context.Background(), data)
		if err != nil {
			return false, false, err
		}
	}
	return false, false, nil
}

func updateReports(coll *mongo.Collection, data []interface{}) (bool, error) {
	for _, report := range data {
		filter := bson.M{"url": report.(models.Report).URL}
		update := bson.M{
			"$set": bson.M{
				"title":     report.(models.Report).Title,
				"startTime": report.(models.Report).StartTime,
				"duration":  report.(models.Report).Duration,
				"speakers":  report.(models.Report).Speakers,
			},
		}
		opts := options.Update().SetUpsert(true)
		_, err := coll.UpdateOne(context.Background(), filter, update, opts)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func deleteReports(coll *mongo.Collection, existingURLS []string) (bool, error) {
	if len(existingURLS) != 0 {
		filter := bson.M{"url": bson.M{"$nin": existingURLS}}
		amount, err := coll.DeleteMany(context.Background(), filter)
		if err != nil {
			return false, fmt.Errorf("failed to delete reports: %w", err)
		}
		if amount.DeletedCount > 0 {
			return true, nil
		}
	}
	return false, nil
}

func (c *Client) SelectReports(coll *mongo.Collection) ([]models.Report, error) {
	cursor, err := coll.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var data []models.Report

	if err = cursor.All(context.Background(), &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) InsertOne(coll *mongo.Collection, data interface{}) error {
	_, err := coll.InsertOne(context.Background(), data)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}
		return err
	}
	return nil
}

func (c *Client) SelectUser(coll *mongo.Collection, tgID int) (models.User, error) {
	var user models.User
	filter := bson.D{{"tgID", tgID}}
	err := coll.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (c *Client) UpdateUserID(coll *mongo.Collection, tgID int, identification string) (bool, error) {
	filter := bson.M{"tgID": tgID}
	update := bson.M{"$set": bson.M{
		"identification": identification,
	}}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) AddUserFavReports(coll *mongo.Collection, tgID int, report models.Report) error {
	filter := bson.M{"tgID": tgID}
	update := bson.M{
		"$addToSet": bson.M{
			"favoriteReports": report,
		},
	}
	_, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveUserFavReport(coll *mongo.Collection, tgID int, reportURL string) error {
	filter := bson.M{"tgID": tgID}
	update := bson.M{"$pull": bson.M{
		"favoriteReports": bson.M{"url": reportURL},
	}}
	_, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) SelectUsers(coll *mongo.Collection) ([]models.User, error) {

	cursor, err := coll.Find(context.Background(), bson.M{})

	if err != nil {
		return nil, err
	}

	var users []models.User

	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Client) SelectReport(coll *mongo.Collection, url string) (models.Report, error) {
	var report models.Report

	filter := bson.D{{"url", url}}

	err := coll.FindOne(context.Background(), filter).Decode(&report)

	if err != nil {
		return models.Report{}, err
	}

	return report, nil
}

func (c *Client) SelectEvaluation(coll *mongo.Collection, tgID int, url string) (models.Evaluation, error) {
	var evaluation models.Evaluation

	filter := bson.D{{"tgID", tgID}, {"url", url}}

	err := coll.FindOne(context.Background(), filter).Decode(&evaluation)

	if err != nil {
		return models.Evaluation{}, err
	}

	return evaluation, nil
}

func (c *Client) SelectEvaluations(coll *mongo.Collection, tgID int) ([]models.Evaluation, error) {

	cursor, err := coll.Find(context.Background(), bson.E{Key: "tgID", Value: tgID})
	if err != nil {
		return nil, err
	}

	var evaluations []models.Evaluation

	if err := cursor.All(context.Background(), &evaluations); err != nil {
		return nil, err
	}

	return evaluations, nil
}

func (c *Client) SelectAllEvaluations(coll *mongo.Collection) ([]models.Evaluation, error) {
	cursor, err := coll.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var evaluations []models.Evaluation

	err = cursor.All(context.Background(), &evaluations)

	if err != nil {
		return nil, err
	}

	return evaluations, nil
}

func (c *Client) UpdateEvaluation(coll *mongo.Collection, tgID int, url string, evaluation models.Evaluation) (bool, error) {
	filter := bson.M{"tgID": tgID, "url": url}
	update := bson.M{
		"$set": bson.M{
			"content":     evaluation.Content,
			"performance": evaluation.Performance,
			"comment":     evaluation.Comment,
		},
	}

	updateResult, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return false, err
	}

	return updateResult.ModifiedCount > 0, nil
}
