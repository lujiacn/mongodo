package mongodo

import (
	"context"
	"fmt"
	"testing"

	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	BaseModel `bson:",inline"`
	Identity  string `bson:"Identity,omitempty"`
	Name      string `bson:",omitempty"`
	Age       int    `bson:",omitempty"`
}

var (
	dbName = "mgofun_test"
	dial   = "mongodb://localhost:27017"
)

func TestTextIndex(t *testing.T) {
	Dial = dial
	DBName = dbName
	// mongo client
	MongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(Dial))
	if err != nil {
		fmt.Println(err)
	}

	MongoDB = MongoClient.Database(DBName)

	ctx := context.Background()
	Client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: Dial})
	DB = Client.Database(DBName)
	if err != nil {
		fmt.Println(err)
	}

	user := new(User)
	_, err = CreateTextIndex(user, []string{"Identity", "Name"})
	fmt.Println("text index", err)
}

func TestIndex(t *testing.T) {
	Dial = dial
	DBName = dbName
	// mongo client
	MongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(Dial))
	if err != nil {
		fmt.Println(err)
	}

	MongoDB = MongoClient.Database(DBName)

	ctx := context.Background()
	Client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: Dial})
	DB = Client.Database(DBName)
	if err != nil {
		fmt.Println(err)
	}

	user := new(User)
	out, err := CreateIndex(user, []string{"Identity"}, true)
	fmt.Println(out, err)
	//user.Name = "Tom"
	//user.Age = 10

	a := new(User)
	a.Identity = "jia"
	do := New(a)
	do.Operator = "Jia"
	do.SaveLog = true
	err = do.Create()
	fmt.Println(err)

	b := new(User)
	b.Identity = "lu"
	do = New(b)
	do.Operator = "Jia"
	do.SaveLog = true
	err = do.Create()
	fmt.Println(err)

	c := new(User)
	c.Identity = "lu"
	do = New(c)
	do.Operator = "Jia"
	do.SaveLog = true
	err = do.Create()
	fmt.Println(err)

	// delete
	d := new(User)
	do = New(d)
	do.Query = bson.M{"Identity": "lu"}
	do.GetByQ()
	do.SaveLog = true
	err = do.Delete()
	fmt.Println(err)

	// create again
	e := new(User)
	do = New(e)
	e.Identity = "lu"
	do.SaveLog = true
	err = do.Create()
	fmt.Println(err)

	result, err := CollectionIndexes(&User{})
	fmt.Println(result, err)
}

func TestCreate(t *testing.T) {
	Dial = dial
	DBName = dbName
	ctx := context.Background()
	//Client, err = qmgo.Open(ctx, &qmgo.Config{Uri: Dial, Database: DBName})
	//ctx := context.Background()
	Client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: Dial})
	DB = Client.Database(DBName)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Client is", Client)

	user := new(User)
	user.Name = "Tom"
	user.Age = 10
	do := New(user)
	do.Operator = "Jia"
	do.SaveLog = true
	do.Create()
	fmt.Println("err", err)
	// test update

	user.Age = 30
	do.Save()
	// soft delte
	do.Delete()

	// get by Id
	newUser := new(User)
	newUser.ID, _ = primitive.ObjectIDFromHex("5f770950d97de5663f329047")
	do = New(newUser)
	err = do.Get()
	fmt.Println("user, err", newUser, err)

	b := new(User)
	do = New(b)
	do.Query = bson.M{"name": "ABC"}
	do.GetByQ()
	fmt.Println("b, err", b, err)
	// test remove
	//do.Remove()

	// removeall

	//record := new(User)
	//do = New(record)
	//do.Query = bson.M{"name": "Tom"}
	//count, err := do.RemoveAll()
	//fmt.Println("removall err, count", err, count)
}
