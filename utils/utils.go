package utils

import (
	"encoding/json"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

type AppConfig struct {
	Server, MongoDBHost, DBUser, DBPwd, Database string
}

var session *mgo.Session
var Config AppConfig
var SignKey []byte
var VerifyKey []byte

func CreateDBSession() *mgo.Session {
	var err error
	dialInfo := mgo.DialInfo{
		Addrs:    []string{Config.MongoDBHost},
		Timeout:  time.Second * 60,
		Database: Config.Database,
	}
	session, err = mgo.DialWithInfo(&dialInfo)
	if err != nil {
		log.Fatal("Error establishing a connection to mongodb server:", err)
		return nil
	}
	return session
}

func StartUp() {
	LoadAppConfig()
	LoadKeys()
}

func LoadAppConfig() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failure to get the current working directory:", err)
		return
	}

	f, err := os.Open(filepath.Join(cwd, "/config.json"))
	if err != nil {
		log.Fatalln("Error while trying to read config file:", err)
	}
	err = json.NewDecoder(f).Decode(&Config)
	if err != nil {
		log.Fatalln("Error decoding configuration file:", err)
	}
}

func LoadKeys() {
	var err error
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failure to get the current working directory:", err)
		return
	}

	SignKey, err = ioutil.ReadFile(filepath.Join(cwd, "/keys/app.rsa"))

	if err != nil {
		log.Fatal("Error reading app.rsa:", err)
		return
	}

	VerifyKey, err = ioutil.ReadFile(filepath.Join(cwd, "/keys/app.rsa.pub"))
	if err != nil {
		log.Fatal("Error reading app.rsa.pub:", err)
		return
	}
}

func AddIndex() {
	messageIndex := mgo.Index{
		Key:        []string{"receiver_id"},
		Background: true,
		Sparse:     true,
	}

	messageCollection := session.DB(Config.Database).C("messages")
	err := messageCollection.EnsureIndex(messageIndex)
	if err != nil {
		log.Println("Error in adding messageIndex:", err)
	}
}
