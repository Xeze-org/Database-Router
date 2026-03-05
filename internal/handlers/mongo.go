package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestMongoConnection returns the live connection status.
func (h *Handler) TestMongoConnection(c *gin.Context) {
	if h.DB.MongoDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MongoDB not enabled"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.DB.MongoDB.Ping(ctx, nil); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "disconnected", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "connected", "database": h.DB.Config.Mongo.Database})
}

// ListMongoDatabases returns all database names.
func (h *Handler) ListMongoDatabases(c *gin.Context) {
	if h.DB.MongoDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MongoDB not enabled"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	databases, err := h.DB.MongoDB.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"databases": databases})
}

// ListMongoCollections returns all collections in a database.
func (h *Handler) ListMongoCollections(c *gin.Context) {
	if h.DB.MongoDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MongoDB not enabled"})
		return
	}
	database := c.Param("database")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collections, err := h.DB.MongoDB.Database(database).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"database": database, "collections": collections})
}

// InsertMongoDocument inserts a JSON body as a document.
func (h *Handler) InsertMongoDocument(c *gin.Context) {
	if h.DB.MongoDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MongoDB not enabled"})
		return
	}
	database := c.Param("database")
	collection := c.Param("collection")

	var document map[string]interface{}
	if err := c.ShouldBindJSON(&document); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.DB.MongoDB.Database(database).Collection(collection).InsertOne(ctx, document)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"database":    database,
		"collection":  collection,
		"inserted_id": result.InsertedID,
	})
}

// FindMongoDocuments returns all documents in a collection.
func (h *Handler) FindMongoDocuments(c *gin.Context) {
	if h.DB.MongoDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MongoDB not enabled"})
		return
	}
	database := c.Param("database")
	collection := c.Param("collection")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := h.DB.MongoDB.Database(database).Collection(collection).Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"database":   database,
		"collection": collection,
		"documents":  results,
		"count":      len(results),
	})
}

// UpdateMongoDocument updates a document by ObjectID.
func (h *Handler) UpdateMongoDocument(c *gin.Context) {
	if h.DB.MongoDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MongoDB not enabled"})
		return
	}
	database := c.Param("database")
	collection := c.Param("collection")
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var update map[string]interface{}
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.DB.MongoDB.Database(database).Collection(collection).UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"database":       database,
		"collection":     collection,
		"matched_count":  result.MatchedCount,
		"modified_count": result.ModifiedCount,
	})
}

// DeleteMongoDocument removes a document by ObjectID.
func (h *Handler) DeleteMongoDocument(c *gin.Context) {
	if h.DB.MongoDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MongoDB not enabled"})
		return
	}
	database := c.Param("database")
	collection := c.Param("collection")
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.DB.MongoDB.Database(database).Collection(collection).DeleteOne(
		ctx,
		bson.M{"_id": objectID},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"database":      database,
		"collection":    collection,
		"deleted_count": result.DeletedCount,
	})
}
