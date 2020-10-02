package mongodo

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/qiniu/qmgo"
	"github.com/revel/revel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	DBName string
	Dial   string
	Client *qmgo.Client
)

// Connect to database and return client
func Connect() {
	var found bool

	Dial = revel.Config.StringDefault("mongodb.dial", "localhost")
	if DBName, found = revel.Config.String("mongodb.name"); !found {
		urls := strings.Split(Dial, "/")
		DBName = urls[len(urls)-1]
	}
	ctx := context.Background()
	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: Dial})
	if err != nil {
		revel.AppLog.Critf("Could not connect to Mongo DB. Error: %s", err)
	}
	Client = client
}

func NewSession() (*qmgo.Session, error) {
	return Client.Session()
}

//MongoController including the mgo session
type MongoController struct {
	MongoSession *qmgo.Session
}

func ControllerInit() {
	revel.InterceptMethod((*MongoController).Begin, revel.BEFORE)
	revel.InterceptMethod((*MongoController).End, revel.FINALLY)
}

//Begin do mongo connection
func (c *MongoController) Begin() revel.Result {
	var err error
	if Client == nil {
		Connect()
	}

	c.MongoSession, err = Client.Session()
	if err != nil {
		return revel.Controller{}.RenderError(err)
	}
	return nil
}

//End close mgo session
func (c *MongoController) End() revel.Result {
	if c.MongoSession != nil {
		c.MongoSession.EndSession(context.Background())
	}
	return nil
}

// ObjectIDBinder do binding
var ObjectIDBinder = revel.Binder{
	// Make a ObjectId from a request containing it in string format.
	Bind: revel.ValueBinder(func(val string, typ reflect.Type) reflect.Value {
		if len(val) == 0 {
			return reflect.Zero(typ)

		}
		if objID, err := primitive.ObjectIDFromHex(val); err == nil {
			return reflect.ValueOf(objID)

		}

		revel.AppLog.Errorf("ObjectIDBinder.Bind - invalid ObjectId!")
		return reflect.Zero(typ)

	}),
	// Turns ObjectId back to hexString for reverse routing
	Unbind: func(output map[string]string, name string, val interface{}) {
		var hexStr string
		hexStr = fmt.Sprintf("%s", val.(primitive.ObjectID).Hex())
		// not sure if this is too carefull but i wouldn't want invalid ObjectIds in my App
		output[name] = hexStr
	},
}
