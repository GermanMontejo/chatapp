package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	Id           bson.ObjectId `bson:"_id" json:"_id"`
	Firstname    string        `json:"firstname"`
	Lastname     string        `json:"lastname"`
	Email        string        `json:"email"`
	Password     string        `json:"password"`
	HashPassword []byte        `json:"hashpassword"`
	Age          string        `json:"age"`
	Address      string        `json:"address"`
	Occupation   string        `json:"occupation"`
	CivilStatus  string        `json:"civil_status"`
	Availability string        `json:"availability"`
	Token        string        `json:"token"`
}

type LoginResponse struct {
	User
	Token string `json:"token"`
}

type Message struct {
	Id         bson.ObjectId `bson:"_id" json:"_id"`
	SenderId   bson.ObjectId `bson:"sender_id" json:"sender_id"`
	ReceiverId bson.ObjectId `bson:"receiver_id" json:"receiver_id"`
	Body       string        `json:"body"`
	Sender     string        `json:"sender"`
	Date       time.Time     `json:"date"`
	MsgType    string        `json:"message_type"`
}
