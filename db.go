package mongodo

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/qiniu/qmgo"
	"github.com/revel/revel"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DBName   string
	Dial     string
	Client   *qmgo.Client // the Client for all connections ?
	DB       *qmgo.Database
	MongoCli *mongo.Client
	MongoDB  *mongo.Database
)

func Init() {
	Connect()
	objID := primitive.NewObjectID()
	revel.TypeBinders[reflect.TypeOf(objID)] = ObjectIDBinder
}

func NewClient(ctx context.Context) (*qmgo.Client, error) {
	return qmgo.NewClient(ctx, &qmgo.Config{Uri: Dial})
}

// Connect to database and return client
func Connect() {
	var found bool
	var err error

	if Dial, found = revel.Config.String("mongodb.dial"); !found {
		revel.AppLog.Crit("Mongodb connection not defined")
	}

	if DBName, found = revel.Config.String("mongodb.name"); !found {
		urls := strings.Split(Dial, "/")
		DBName = urls[len(urls)-1]
	}

	// qmgo client
	ctx := context.Background()
	//Client, err = qmgo.Open(ctx, &qmgo.Config{Uri: Dial, Database: DBName})
	//ctx := context.Background()
	Client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: Dial})
	if err != nil {
		revel.AppLog.Errorf("Could not connect to Mongo DB. Error: %s", err)
	}

	DB = Client.Database(DBName)
	if err != nil {
		revel.AppLog.Errorf("Could not connect to Mongo DB. Error: %s", err)
	}

	// mongo client
	MongoCli, err := mongo.Connect(ctx, options.Client().ApplyURI(Dial))
	if err != nil {
		revel.AppLog.Errorf("Could not connect to Mongo DB. Error: %s", err)
	}

	MongoDB = MongoCli.Database(DBName)

}

//MongoController including the mgo session
//type MongoController struct {
//MongoCli *qmgo.QmgoClient
//}

//func ControllerInit() {
//revel.InterceptMethod((*MongoController).Begin, revel.BEFORE)
//}

//Begin do mongo connection
//func (c *MongoController) Begin() revel.Result {
//if Client == nil {
//Connect()
//}
//return nil
//}

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
