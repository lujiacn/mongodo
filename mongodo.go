package mongodo

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChangeLog Collaction Name
const CHANGELOG = "ChangeLog"

const (
	UPDATE = "UPDATE"
	DELETE = "DELETE" // soft delete
	REMOVE = "REMOVE" // hard delete
	CREATE = "CREATE" // hard delete
)

type Do struct {
	model    interface{}
	Ctx      context.Context
	Query    bson.M
	Client   *qmgo.Client
	Coll     *qmgo.Collection
	Sort     []string
	Skip     int64
	Limit    int64
	Select   []string
	Operator string
	Reason   string
	SaveLog  bool // default is false
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
	return &Do{model: model, Coll: coll, Ctx: context.Background()}
}

// NewWithC parse record from different collection
func NewWithC(model interface{}, cName string) *Do {
	coll := DB.Collection(cName)
	return &Do{model: model, Coll: coll, Ctx: context.Background()}
}

// NewCtx with Ctx input
func NewCtx(ctx context.Context, model interface{}) *Do {
	colName := getModelName(model)
	coll := DB.Collection(colName)
	return &Do{model: model, Coll: coll, Ctx: ctx}
}

// Create will generate ID for model
func (m *Do) Create() error {
	timeNow := time.Now()
	id := reflect.ValueOf(m.model).Elem().FieldByName("ID")
	x := reflect.ValueOf(m.model).Elem().FieldByName("CreatedAt")
	x.Set(reflect.ValueOf(timeNow))
	by := reflect.ValueOf(m.model).Elem().FieldByName("CreatedBy")
	by.Set(reflect.ValueOf(m.Operator))
	l := reflect.ValueOf(m.model).Elem().FieldByName("LatestTime")
	l.Set(reflect.ValueOf(timeNow))

	isRemoved := reflect.ValueOf(m.model).Elem().FieldByName("IsRemoved")
	f := false
	isRemoved.Set(reflect.ValueOf(&f))
	// Important: make sure the sessCtx used in every operation in the whole transaction
	// start transaction
	if result, err := m.Coll.InsertOne(m.Ctx, m.model); err != nil {
		return err
	} else {
		id.Set(reflect.ValueOf(result.InsertedID))
	}

	if !m.SaveLog {
		return nil
	}
	time.Sleep(1 * time.Second)
	if err := m.saveLog(m.Ctx, CREATE); err != nil {
		return err
	}
	return nil
}

// Save update one record according ID
func (m *Do) Save() error {
	timeNow := time.Now()
	id := reflect.ValueOf(m.model).Elem().FieldByName("ID")
	x := reflect.ValueOf(m.model).Elem().FieldByName("UpdatedAt")
	x.Set(reflect.ValueOf(timeNow))
	by := reflect.ValueOf(m.model).Elem().FieldByName("UpdatedBy")
	by.Set(reflect.ValueOf(m.Operator))
	l := reflect.ValueOf(m.model).Elem().FieldByName("LatestTime")
	l.Set(reflect.ValueOf(timeNow))

	if err := m.Coll.UpdateOne(m.Ctx,
		bson.M{"_id": id.Interface()}, bson.M{"$set": m.model}); err != nil {
		return err
	}

	if !m.SaveLog {
		return nil
	}

	if err := m.saveLog(m.Ctx, UPDATE); err != nil {
		return err
	}

	return nil
}

// Erase is alias for Remove
func (m *Do) Erase() error {
	return m.Remove()
}

// Remove is hard delete
func (m *Do) Remove() error {
	id := reflect.ValueOf(m.model).Elem().FieldByName("ID")
	if err := m.Coll.Remove(m.Ctx, bson.M{"_id": id.Interface()}); err != nil {
		return err
	}

	if !m.SaveLog {
		return nil
	}

	if err := m.saveLog(m.Ctx, REMOVE); err != nil {
		return err
	}

	return nil
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

	if err := m.Coll.UpdateOne(m.Ctx, bson.M{"_id": id.Interface()}, bson.M{"$set": m.model}); err != nil {
		return err
	}

	if !m.SaveLog {
		return nil
	}

	if err := m.saveLog(m.Ctx, DELETE); err != nil {
		return err
	}

	return nil
}

// -----  common functions -----

// findQ is basic Q for all query, added IsRemoved : false
func (m *Do) findQ() qmgo.QueryI {
	if m.Query == nil {
		m.Query = bson.M{}
	}
	m.Query["IsRemoved"] = bson.M{"$ne": true}

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

//
//func (m *Do)GetByQAndUpdate() error {
//err :=
//}

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

func (m *Do) EraseAll() (int64, error) {
	return m.RemoveAll()
}

// RemoveAll is hardDelete
func (m *Do) RemoveAll() (int64, error) {
	if m.Query == nil {
		return 0, errors.New("Cannot remove without condition")
	}

	result, err := m.Coll.RemoveAll(context.Background(), m.Query)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

//saveLog just copy a record to Changlog
func (m *Do) saveLog(ctx context.Context, operation string) error {
	id := reflect.ValueOf(m.model).Elem().FieldByName("ID")

	cl := new(ChangeLog)
	cl.ChangeReason = m.Reason
	cl.Operation = operation
	cl.ModelObjectID = id.Interface().(primitive.ObjectID)
	cl.ModelName = getModelName(m.model)
	cl.ModelValue = m.model
	cl.Operator = m.Operator
	do := NewCtx(ctx, cl)
	err := do.Create()
	return err
}

// FetchByQAndDelete find One record according to Query and mark as IsRemoved
func (m *Do) FetchByQAndDelete() error {
	colName := getModelName(m.model)
	coll := MongoDB.Collection(colName)
	if m.Query == nil {
		m.Query = bson.M{}
	}

	m.Query["IsRemoved"] = bson.M{"$ne": true}

	result := coll.FindOneAndUpdate(context.Background(), m.Query, bson.M{"$set": bson.M{"IsRemoved": true}}, &options.FindOneAndUpdateOptions{})
	if result.Err() != nil {
		return result.Err()
	}
	decodeErr := result.Decode(m.model)
	if decodeErr != nil {
		return decodeErr
	}
	return nil
}
