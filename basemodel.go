package mongodo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BaseModel to be emmbered to other struct as audit trail perpurse
type BaseModel struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt  time.Time          `bson:"CreatedAt,omitempty"`
	CreatedBy  string             `bson:"CreatedBy,omitempty"`
	UpdatedAt  time.Time          `bson:"UpdatedAt,omitempty"`
	UpdatedBy  string             `bson:"UpdatedBy,omitempty"`
	IsRemoved  *bool              `bson:"IsRemoved,omitempty"`
	RemovedAt  time.Time          `bson:"RemovedAt,omitempty"`
	RemovedBy  string             `bson:"RemovedBy,omitempty"`
	IsLocked   *bool              `bson:"IsLocked,omitempty"`
	LatestTime time.Time          `bson:"LatestTime,omitempty"`
}

//ChangeLog
type ChangeLog struct {
	BaseModel     `bson:",inline"`
	ModelObjectID primitive.ObjectID `bson:"ModelObjectID,omitempty"`
	ModelName     string             `bson:"ModelName,omitempty"`
	ModelValue    interface{}        `bson:"ModelValue,omitempty"`
	Operation     string             `bson:"Operation,omitempty"`
	ChangeReason  string             `bson:"ChangeReason,omitempty"`
	Operator      string             `bson:"Operator,omitempty"`
}

// VisitLog
type VisitLog struct {
	BaseModel `bson:",inline"`
	IP        string `bson:"IP,omitempty"`
	Identity  string `bson:"Identity,omitempty"`
	Locale    string `bson:"Locale,omitempty"`
	Url       string `bson:"Url,omitempty"`
}
