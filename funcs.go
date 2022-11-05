package mongodo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateIndexSpecPartial set partiaFilter by user
func CreateIndexSpecPartial(model interface{}, keys []string, unique bool, partialFilter bson.M) (string, error) {
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
			PartialFilterExpression: partialFilter,
			Unique:                  &unique,
		},
	}

	out, err := indexView.CreateOne(context.Background(), indexModel, options.CreateIndexes())
	if err != nil {
		return "", err
	}
	return out, nil
}

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

// All index without PartialFilterExperession IsRemoved = false
// CreateIndex with keys and unique (true or false)
func CreateIndexNoPartial(model interface{}, keys []string, unique bool) (string, error) {
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
			Unique: &unique,
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
		//Options: &options.IndexOptions{
		//PartialFilterExpression: bson.M{"IsRemoved": false},
		//},
	}

	out, err := indexView.CreateOne(context.Background(), indexModel, options.CreateIndexes())
	if err != nil {
		return "", err
	}
	return out, nil
}

// CreateExpireIndex create expire index for single field
func CreateExpireIndex(model interface{}, fieldName string, expireAfter int32) (string, error) {
	colName := getModelName(model)
	coll := MongoDB.Collection(colName)
	indexView := coll.Indexes()
	indexModel := mongo.IndexModel{
		Keys:    bson.M{fieldName: 1},
		Options: options.Index().SetExpireAfterSeconds(expireAfter),
	}

	out, err := indexView.CreateOne(context.Background(), indexModel, options.CreateIndexes())
	return out, err
}

// IsDup check whether the err is Duplicate error
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

// CollectionIndexes listout collection indexes
func CollectionIndexes(model interface{}) ([]string, error) {
	colName := getModelName(model)
	coll := MongoDB.Collection(colName)
	indexView := coll.Indexes()
	opts := options.ListIndexes().SetMaxTime(2 * time.Second)

	var results []bson.M

	cursor, err := indexView.List(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	var output []string
	for _, item := range results {
		output = append(output, item["name"].(string))
	}
	return output, nil
}

// Drop One Index from collection with indexName
func DropOneIndex(model interface{}, name string) error {
	colName := getModelName(model)
	coll := MongoDB.Collection(colName)
	indexView := coll.Indexes()
	_, err := indexView.DropOne(context.Background(), name)
	return err
}
