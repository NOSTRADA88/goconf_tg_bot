// Package mongodb provides a client for interacting with a MongoDB database.
package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/NOSTRADA88/telegram-bot-go/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ctx is a global context used for MongoDB operations.
var ctx context.Context

// Client is a struct that wraps a MongoDB client.
type Client struct {
	mongo *mongo.Client
}

// DataManipulator is an interface that defines methods for manipulating data in the database.
type DataManipulator interface {
	ReportManipulator
	UserManipulator
	EvaluationManipulator
	Init(context context.Context) error
	Collection(collection string) *mongo.Collection
	InsertOne(coll *mongo.Collection, data interface{}) error
}

// ReportManipulator is an interface that defines methods for manipulating report data.
type ReportManipulator interface {
	InsertMany(coll *mongo.Collection, data []interface{}) (bool, bool, error)
	SelectReport(coll *mongo.Collection, url string) (models.Report, error)
	SelectReports(coll *mongo.Collection) ([]models.Report, error)
}

// UserManipulator is an interface that defines methods for manipulating user data.
type UserManipulator interface {
	SelectUser(coll *mongo.Collection, tgID int) (models.User, error)
	SelectUsers(coll *mongo.Collection) ([]models.User, error)
	UpdateUserID(coll *mongo.Collection, tgID int, identification string) (bool, error)
	AddUserFavReports(coll *mongo.Collection, tgID int, report models.Report) error
	RemoveUserFavReport(coll *mongo.Collection, tgID int, reportURL string) error
}

// EvaluationManipulator is an interface that defines methods for manipulating evaluation data.
type EvaluationManipulator interface {
	SelectEvaluation(coll *mongo.Collection, tgID int, url string) (bool, models.Evaluation, error)
	SelectEvaluations(coll *mongo.Collection, tgID int) ([]models.Evaluation, error)
	SelectAllEvaluations(coll *mongo.Collection) ([]models.Evaluation, error)
	UpdateEvaluation(coll *mongo.Collection, tgID int, url string, evaluation models.Evaluation) (bool, error)
	DeleteEvaluation(coll *mongo.Collection, tgID int, url string) (bool, error)
}

// New creates a new MongoDB client and returns it.
func New(host string, port int, user, password string) (*Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%v", user, password, host, port)))
	if err != nil {
		return nil, err
	}
	return &Client{mongo: client}, nil
}

// Close disconnects the MongoDB client.
func (c *Client) Close() error {
	return c.mongo.Disconnect(ctx)
}

// Init initializes the MongoDB client with a given context and ensures uniqueness of user IDs and evaluations.
func (c *Client) Init(context context.Context) error {
	ctx = context
	if err := c.ensureUserTgIDUnique(); err != nil {
		return err
	}
	return c.ensureEvaluationTgIDAndURLUnique()
}

// ensureUserTgIDUnique ensures that the user ID is unique in the database.
func (c *Client) ensureUserTgIDUnique() error {
	coll := c.Collection("user")
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"tgID": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	return err
}

// ensureEvaluationTgIDAndURLUnique ensures that the combination of user ID and URL is unique in the database.
func (c *Client) ensureEvaluationTgIDAndURLUnique() error {
	coll := c.Collection("evaluation")
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "tgID", Value: 1},
			{Key: "url", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	return err
}

// Collection returns a MongoDB collection with the given name.
func (c *Client) Collection(collection string) *mongo.Collection {
	return c.mongo.Database("tg-bot").Collection(collection)
}

// InsertMany inserts multiple documents into a collection.
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
		_, err := coll.InsertMany(ctx, data)
		if err != nil {
			return false, false, err
		}
	}
	return false, false, nil
}

// updateReports updates existing reports in the database.
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
		_, err := coll.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func deleteReports(coll *mongo.Collection, existingURLS []string) (bool, error) {
	if len(existingURLS) != 0 {
		filter := bson.M{"url": bson.M{"$nin": existingURLS}}
		amount, err := coll.DeleteMany(ctx, filter)
		if err != nil {
			return false, fmt.Errorf("failed to delete reports: %w", err)
		}
		if amount.DeletedCount > 0 {
			return true, nil
		}
	}
	return false, nil
}

// SelectReports selects all reports from a collection.
func (c *Client) SelectReports(coll *mongo.Collection) ([]models.Report, error) {
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var data []models.Report

	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// InsertOne inserts a single document into a collection.
func (c *Client) InsertOne(coll *mongo.Collection, data interface{}) error {
	_, err := coll.InsertOne(ctx, data)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}
		return err
	}
	return nil
}

// SelectUser selects a user from a collection by their Telegram ID.
func (c *Client) SelectUser(coll *mongo.Collection, tgID int) (models.User, error) {
	var user models.User
	filter := bson.D{{"tgID", tgID}}
	err := coll.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// UpdateUserID updates a user's identification in the database.
func (c *Client) UpdateUserID(coll *mongo.Collection, tgID int, identification string) (bool, error) {
	filter := bson.M{"tgID": tgID}
	update := bson.M{"$set": bson.M{
		"identification": identification,
	}}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return false, err
	}
	return true, nil
}

// AddUserFavReports adds a report to a user's list of favorite reports.
func (c *Client) AddUserFavReports(coll *mongo.Collection, tgID int, report models.Report) error {
	filter := bson.M{"tgID": tgID}
	update := bson.M{
		"$addToSet": bson.M{
			"favoriteReports": report,
		},
	}
	_, err := coll.UpdateOne(ctx, filter, update)
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
	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) SelectUsers(coll *mongo.Collection) ([]models.User, error) {

	cursor, err := coll.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	var users []models.User

	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Client) SelectReport(coll *mongo.Collection, url string) (models.Report, error) {
	var report models.Report

	filter := bson.D{{"url", url}}

	err := coll.FindOne(ctx, filter).Decode(&report)

	if err != nil {
		return models.Report{}, err
	}

	return report, nil
}

// SelectEvaluation selects an evaluation from a collection by the user's Telegram ID and the report URL.
func (c *Client) SelectEvaluation(coll *mongo.Collection, tgID int, url string) (bool, models.Evaluation, error) {
	var evaluation models.Evaluation

	filter := bson.D{{"tgID", tgID}, {"url", url}}

	err := coll.FindOne(ctx, filter).Decode(&evaluation)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, models.Evaluation{}, nil
		}
		return false, models.Evaluation{}, err
	}

	return true, evaluation, nil
}

// SelectEvaluations selects all evaluations from a collection by the user's Telegram ID.
func (c *Client) SelectEvaluations(coll *mongo.Collection, tgID int) ([]models.Evaluation, error) {

	cursor, err := coll.Find(ctx, bson.M{"tgID": tgID})

	if err != nil {
		return nil, err
	}

	var evaluations []models.Evaluation

	if err = cursor.All(ctx, &evaluations); err != nil {
		return nil, err
	}

	return evaluations, nil
}

// SelectAllEvaluations selects all evaluations from a collection.
func (c *Client) SelectAllEvaluations(coll *mongo.Collection) ([]models.Evaluation, error) {
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var evaluations []models.Evaluation

	err = cursor.All(ctx, &evaluations)

	if err != nil {
		return nil, err
	}

	return evaluations, nil
}

// UpdateEvaluation updates an evaluation in the database.
func (c *Client) UpdateEvaluation(coll *mongo.Collection, tgID int, url string, evaluation models.Evaluation) (bool, error) {
	filter := bson.M{"tgID": tgID, "url": url}
	update := bson.M{
		"$set": bson.M{
			"content":     evaluation.Content,
			"performance": evaluation.Performance,
			"comment":     evaluation.Comment,
		},
	}

	updateResult, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}

	return updateResult.ModifiedCount > 0, nil
}

// DeleteEvaluation deletes an evaluation from the database.
func (c *Client) DeleteEvaluation(coll *mongo.Collection, tgID int, url string) (bool, error) {

	deleted, err := coll.DeleteOne(ctx, bson.M{"tgID": tgID, "url": url})

	if err != nil {
		return false, err
	}

	return deleted.DeletedCount > 0, nil
}
