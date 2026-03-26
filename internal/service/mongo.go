package service

import (
	"context"
	"fmt"
	"time"

	"db-router/internal/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mongoService struct {
	db *db.Manager
}

// NewMongoService constructs a MongoService backed by the given db.Manager.
func NewMongoService(m *db.Manager) MongoService {
	return &mongoService{db: m}
}

func (s *mongoService) ListDatabases(ctx context.Context) ([]string, error) {
	if s.db.MongoDB == nil {
		return nil, ErrNotEnabled("MongoDB")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return s.db.MongoDB.ListDatabaseNames(ctx, bson.M{})
}

func (s *mongoService) ListCollections(ctx context.Context, database string) ([]string, error) {
	if s.db.MongoDB == nil {
		return nil, ErrNotEnabled("MongoDB")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return s.db.MongoDB.Database(database).ListCollectionNames(ctx, bson.M{})
}

func (s *mongoService) InsertDocument(ctx context.Context, database, collection string, document Row) (string, error) {
	if s.db.MongoDB == nil {
		return "", ErrNotEnabled("MongoDB")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := s.db.MongoDB.Database(database).Collection(collection).InsertOne(ctx, document)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", result.InsertedID), nil
}

func (s *mongoService) FindDocuments(ctx context.Context, database, collection string) ([]Row, error) {
	if s.db.MongoDB == nil {
		return nil, ErrNotEnabled("MongoDB")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := s.db.MongoDB.Database(database).Collection(collection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	results := make([]Row, 0, len(docs))
	for _, doc := range docs {
		row := make(Row, len(doc))
		for k, v := range doc {
			row[k] = v
		}
		results = append(results, row)
	}
	return results, nil
}

func (s *mongoService) UpdateDocument(ctx context.Context, database, collection, id string, update Row) (int64, int64, error) {
	if s.db.MongoDB == nil {
		return 0, 0, ErrNotEnabled("MongoDB")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid ID format: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := s.db.MongoDB.Database(database).Collection(collection).UpdateOne(
		ctx, bson.M{"_id": objectID}, bson.M{"$set": update})
	if err != nil {
		return 0, 0, err
	}
	return result.MatchedCount, result.ModifiedCount, nil
}

func (s *mongoService) DeleteDocument(ctx context.Context, database, collection, id string) (int64, error) {
	if s.db.MongoDB == nil {
		return 0, ErrNotEnabled("MongoDB")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := s.db.MongoDB.Database(database).Collection(collection).DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (s *mongoService) TestConnection(ctx context.Context) (string, error) {
	if s.db.MongoDB == nil {
		return "", ErrNotEnabled("MongoDB")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.db.MongoDB.Ping(ctx, nil); err != nil {
		return "", err
	}
	return s.db.Config.Mongo.Database, nil
}
