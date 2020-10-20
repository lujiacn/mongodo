package mongodo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// All index with PartialFilterExperession IsRemoved = false
// CreateIndex with keys and unique (true or false)
func CreateIndex(model interface{}, keys []string, unique bool) (string, error) {
	colName := getModelName(model)
	coll := MongoDB.Collection(colName)
	indexView := coll.Indexes()
	bKeys := bson.D{}
	for _, k := range keys {
		bKeys = append(bKeys, primitive.E{Key: k, Value: 1})
	}

	indexModel := mongo.IndexModel{
		Keys: bKeys,
		Options: &options.IndexOptions{
			PartialFilterExpression: bson.M{"IsRemoved": false},
			Unique:                  &unique,
		},
	}

	out, err := indexView.CreateOne(context.Background(), indexModel, options.CreateIndexes())
	if err != nil {
		return "", err
	}
	return out, nil
}

func CreateTextIndex(model interface{}, keys []string) (string, error) {
	colName := getModelName(model)
	coll := MongoDB.Collection(colName)
	indexView := coll.Indexes()
	bKeys := bson.D{}
	for _, k := range keys {
		bKeys = append(bKeys, primitive.E{Key: k, Value: "text"})
	}

	indexModel := mongo.IndexModel{
		Keys: bKeys,
		Options: &options.IndexOptions{
			PartialFilterExpression: bson.M{"IsRemoved": false},
		},
	}

	out, err := indexView.CreateOne(context.Background(), indexModel, options.CreateIndexes())
	if err != nil {
		return "", err
	}
	return out, nil
}

func IsDup(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}
