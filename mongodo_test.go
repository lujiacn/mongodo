package mongodo

import (
	"context"
	"fmt"
	"testing"

	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	BaseModel `bson:",inline"`
	Name      string `bson:"name,omitempty"`
	Age       int    `bson:"age,omitempty"`
}

var (
	dbName = "mgofun_test"
	dial   = "mongodb://localhost:27017"
)

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

	record := new(User)
	do = New(record)
	do.Query = bson.M{"name": "Tom"}
	count, err := do.RemoveAll()
	fmt.Println("removall err, count", err, count)
}
