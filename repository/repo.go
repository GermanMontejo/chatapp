package repository

import (
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"

	. "github.com/GermanMontejo/chatapp/model"
	. "github.com/GermanMontejo/chatapp/utils"
)

type Context struct {
	Session *mgo.Session
}

type Repository struct {
	Category *mgo.Collection
}

var session *mgo.Session

func (c *Context) Close() {
	defer c.Session.Close()
}

func GetSession() *mgo.Session {
	session = CreateDBSession()
	AddIndex()
	return session
}

func (r *Repository) ValidateUser(email string, hpass []byte) (User, bool) {
	user := User{}
	err := r.Category.Find(bson.M{"email": email}).One(&user)

	if err != nil {
		log.Println("User data is not stored in the database. Please sign up...")
		return user, false
	}

	err = bcrypt.CompareHashAndPassword(user.HashPassword, hpass)

	if err != nil {
		log.Println("Error comparing hash password:", err)
		return user, false
	}

	if user.Email != "" {
		return user, true
	}

	return user, false
}

func (r *Repository) AddUser(u *User) error {
	log.Println("AddUser")
	u.Id = bson.NewObjectId()
	encPass, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Bcrypt password did not work, resolving issue by manually converting string password to []byte.")
		u.HashPassword = []byte(u.Password)
	} else {
		u.HashPassword = encPass
	}
	u.Password = ""
	err = r.Category.Insert(&u)
	if err != nil {
		log.Println("Error adding a new user to the database:", err)
		return err
	}
	return nil
}

func (r *Repository) UpdateUser(id bson.ObjectId, u *User) error {
	err := r.Category.Update(bson.M{"_id": id}, &u)
	if err != nil {
		log.Println("Error while trying to update a user document:", err)
	}
	return err
}

func (r *Repository) GetUser(id bson.ObjectId) (u User) {
	r.Category.Find(bson.M{"_id": id}).One(&u)
	return u
}

func (r *Repository) GetUserByEmail(email string) (u User) {
	r.Category.Find(bson.M{"email": email}).One(&u)
	return u
}

func (r *Repository) GetAllUsers() []User {
	var users []User
	var user User
	iter := r.Category.Find(nil).Iter()
	for iter.Next(&user) {
		users = append(users, user)
	}
	return users
}

func (r *Repository) RemoveUser(id string) error {
	err := r.Category.RemoveId(bson.ObjectIdHex(id))
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) SaveMessage(msg *Message) error {
	err := r.Category.Insert(&msg)
	log.Println("Saved message:", msg)
	if err != nil {
		log.Println("Error saving message:", err)
		return err
	}
	return nil
}

func (r *Repository) GetMessages(id bson.ObjectId) []Message {
	log.Println("GetMessages")
	log.Println("id:", bson.M{"receiver_id": id})
	iter := r.Category.Find(bson.M{"receiver_id": id}).Iter()
	result := Message{}
	messages := []Message{}

	for iter.Next(&result) {
		log.Println(result)
		messages = append(messages, result)
	}

	return messages
}
