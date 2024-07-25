package db

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   int64              `bson:"user_id"`
	Expenses []Expense          `bson:"expenses"`
}

type Expense struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Amount     float64            `bson:"amount"`
	Date       string             `bson:"date"`
	Note       string             `bson:"note"`
	CategoryID primitive.ObjectID `bson:"category_id"`
}

type Category struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	UserID int64              `bson:"user_id"`
	Name   string             `bson:"name"`
	Limit  float64            `bson:"limit"`
}


