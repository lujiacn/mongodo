package mongodo

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/qiniu/qmgo"
	//"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson"
)

const LogColl string = "ChangeLog"

type Do struct {
	model    interface{}
	Query    bson.M
	Coll     *qmgo.Collection
	Sort     []string
	Skip     int64
	Limit    int64
	Select   []string
	Operator string
	Reason   string
}

//getModelName reflect string name from model
func getModelName(m interface{}) string {
	var c string
	switch m.(type) {
	case string:
		c = m.(string)
	default:
		typ := reflect.TypeOf(m)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		c = typ.Name()
	}
	return c
}

// New using mongodo Client
func New(model interface{}) *Do {
	colName := getModelName(model)
	coll := DB.Collection(colName)
	return &Do{model: model, Coll: coll}
}

// Create will generate ID for model
func (m *Do) Create() error {
	var err error
	id := reflect.ValueOf(m.model).Elem().FieldByName("ID")
	x := reflect.ValueOf(m.model).Elem().FieldByName("CreatedAt")
	x.Set(reflect.ValueOf(time.Now()))
	by := reflect.ValueOf(m.model).Elem().FieldByName("CreatedBy")
	by.Set(reflect.ValueOf(m.Operator))

	isRemoved := reflect.ValueOf(m.model).Elem().FieldByName("IsRemoved")
	f := false
	isRemoved.Set(reflect.ValueOf(&f))

	result, err := m.Coll.InsertOne(context.Background(), m.model)
	id.Set(reflect.ValueOf(result.InsertedID))
	return err
}

// Save update one record according ID
func (m *Do) Save() error {
	id := reflect.ValueOf(m.model).Elem().FieldByName("ID")
	x := reflect.ValueOf(m.model).Elem().FieldByName("UpdatedAt")
	x.Set(reflect.ValueOf(time.Now()))
	by := reflect.ValueOf(m.model).Elem().FieldByName("UpdatedBy")
	by.Set(reflect.ValueOf(m.Operator))

	err := m.Coll.UpdateOne(context.Background(), bson.M{"_id": id.Interface()}, bson.M{"$set": m.model})

	return err
}

// Remove is hard delete
func (m *Do) Remove() error {
	id := reflect.ValueOf(m.model).Elem().FieldByName("ID")
	err := m.Coll.Remove(context.Background(), bson.M{"_id": id.Interface()})
	return err
}

// Delete is soft delete
func (m *Do) Delete() error {
	id := reflect.ValueOf(m.model).Elem().FieldByName("ID")
	x := reflect.ValueOf(m.model).Elem().FieldByName("RemovedAt")
	x.Set(reflect.ValueOf(time.Now()))
	by := reflect.ValueOf(m.model).Elem().FieldByName("RemovedBy")
	by.Set(reflect.ValueOf(m.Operator))
	removed := reflect.ValueOf(m.model).Elem().FieldByName("IsRemoved")
	f := true
	removed.Set(reflect.ValueOf(&f))

	err := m.Coll.UpdateOne(context.Background(), bson.M{"_id": id.Interface()}, bson.M{"$set": m.model})
	return err
}

// -----  common functions -----

// findQ is basic Q for all query, added IsRemoved : false
func (m *Do) findQ() qmgo.QueryI {
	if m.Query != nil {
		m.Query["IsRemoved"] = false
	} else {
		m.Query = bson.M{"IsRemoved": false}
	}

	q := m.Coll.Find(context.Background(), m.Query)
	//sort
	if m.Sort != nil {
		q = q.Sort(m.Sort...)
	}

	//skip
	if m.Skip != 0 {
		q = q.Skip(m.Skip)
	}

	//limit
	if m.Limit != 0 {
		q = q.Limit(m.Limit)
	}

	//Select
	if m.Select != nil {
		sCols := bson.M{}
		for _, v := range m.Select {
			if strings.HasPrefix(v, "-") {
				t := v[1 : len(v)-1]
				sCols[t] = -1
			} else {
				sCols[v] = 1
			}
		}
		q = q.Select(sCols)
	}

	return q
}

func (m *Do) findByIdQ() qmgo.QueryI {
	id := reflect.ValueOf(m.model).Elem().FieldByName("ID").Interface()
	m.Query = bson.M{"_id": id}
	m.findQ()
	return m.findQ()
}

// Get, if ID exist by ID, or query
func (m *Do) Get() error {
	q := m.findByIdQ()
	err := q.One(m.model)
	return err
}

// GetByQ find One record according to Query
func (m *Do) GetByQ() error {
	err := m.findQ().One(m.model)
	return err
}

// Fetch same as Get, but bind to another struct (uses for model name diff)
func (m *Do) Fetch(record interface{}) error {
	err := m.findByIdQ().One(record)
	return err
}

// FetchByQ find One record according to Query
func (m *Do) FetchByQ(record interface{}) error {
	err := m.findQ().One(record)
	return err
}

func (m *Do) FindAll(i interface{}) error {
	err := m.findQ().All(i)
	return err
}

func (m *Do) Count() (int64, error) {
	return m.findQ().Count()
}

func (m *Do) Distinct(key string, i interface{}) error {
	err := m.findQ().Distinct(key, i)
	return err
}

// RemoveAll is hardDelete
func (m *Do) RemoveAll() (int64, error) {
	if m.Query == nil {
		return 0, errors.New("Cannot remove without condition")
	}

	m.Query = bson.M{"IsRemoved": false}
	result, err := m.Coll.RemoveAll(context.Background(), m.Query)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}
