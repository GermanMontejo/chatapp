package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	. "github.com/GermanMontejo/chatapp/model"
	. "github.com/GermanMontejo/chatapp/repository"
	. "github.com/GermanMontejo/chatapp/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/shijuvar/go-web/taskmanager/Godeps/_workspace/src/golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

var context *Context

const (
	userCollection    = "users"
	messageCollection = "messages"
	msgDelivered      = "DELIVERED"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := jwt.ParseFromRequest(r, func(token *jwt.Token) (interface{}, error) {
			return VerifyKey, nil
		})
		if err != nil {

			// We check the email and password submitted. If they match one of the user
			// credentials in our database and the token error they have is related to
			// expired token, then we issue them a new token.

			user := User{}
			repo := GetRepoInstance(userCollection)
			_ = json.NewDecoder(r.Body).Decode(&user)
			_, isValid := repo.ValidateUser(user.Email, []byte(user.Password))

			if isValid && strings.Contains(err.Error(), "token is expired") {
				log.Println("token is expired")
				// issue a new token, since this user's token has already expired.
				t := GenerateJWT(&user)

				user = repo.GetUserByEmail(user.Email)
				user.Token = t
				repo.UpdateUser(user.Id, &user)

				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Add("Content-Type", "application/json")
				newTokenMsg := `{"message":"Your token has expired. Please use this new token:", "token":` + t

				bm := composeJsonResponseForNonStructTypes(newTokenMsg)
				w.Write(bm)
				return
			}

			log.Println("Error while parsing token from request:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "application/json")
			bm := composeJsonResponseForNonStructTypes(err.Error())
			w.Write(bm)
			return
		}
		if token.Valid {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "application/json")
			bm := composeJsonResponseForNonStructTypes("Invalid token")
			w.Write(bm)
		}
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LoginHandler")
	user := User{}
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Println("There is an error while decoding response body into the User struct:", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Add("Content-Type", "applicaton/json")
		bm := composeJsonResponseForNonStructTypes(err.Error())
		w.Write(bm)
	}

	repo := GetRepoInstance(userCollection)
	defer context.Close()
	hpass := []byte(user.Password)

	if err != nil {
		log.Println("Error encrypting password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("Content-Type", "applicaton/json")
		bm := composeJsonResponseForNonStructTypes(err.Error())
		w.Write(bm)
	}

	_, isCreated := repo.ValidateUser(user.Email, hpass)

	if isCreated {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "applicaton/json")
		bm := composeJsonResponseForNonStructTypes("Login success!")
		w.Write(bm)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	w.Header().Add("Content-Type", "applicaton/json")
	bm := composeJsonResponseForNonStructTypes("Record not found! You should sign up.")
	w.Write(bm)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("SignupHandler")
	user := User{}
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Println("Error while decoding request body:", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		bm := composeJsonResponseForNonStructTypes(err.Error())
		w.Write(bm)
	}

	repo := GetRepoInstance(userCollection)
	defer context.Close()
	hpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		log.Println("Error encrypting password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("Content-Type", "applicaton/json")
		bm := composeJsonResponseForNonStructTypes(err.Error())
		w.Write(bm)
	}

	_, isCreated := repo.ValidateUser(user.Email, hpass)

	if isCreated {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "applicaton/json")
		bm := composeJsonResponseForNonStructTypes("There is already an account with this email. Try logging in...")
		w.Write(bm)
		return
	}

	token := GenerateJWT(&user)

	user.Token = token
	err = repo.AddUser(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Add("Content-Type", "applicaton/json")
		bm := composeJsonResponseForNonStructTypes("Error while saving user data to database:" + err.Error())
		w.Write(bm)
		return
	}

	loginResponse := LoginResponse{
		User:  user,
		Token: token,
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Content-Type", "applicaton/json")
	j, err := json.MarshalIndent(&loginResponse, "", "\t")
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		bm := composeJsonResponseForNonStructTypes(err.Error())
		w.Write(bm)
		return
	}
	w.Write(j)
}

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	repo := GetRepoInstance(userCollection)
	users := repo.GetAllUsers()
	j, err := json.Marshal(&users)
	if err != nil {
		log.Println("Error marshaling users data:", err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DeleteUserHandler")
	values := r.URL.Query()
	ids := values["id"]
	id := ids[0]

	if id == "" {
		log.Println("empty id")
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		bm := composeJsonResponseForNonStructTypes("The client has sent an empty id. Please send a valid one.")
		w.Write(bm)
		return
	}

	repo := GetRepoInstance(userCollection)
	err := repo.RemoveUser(id)
	if err != nil {
		log.Println("repo.removeUser")
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		bm := composeJsonResponseForNonStructTypes(err.Error())
		w.Write(bm)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	bm := composeJsonResponseForNonStructTypes("User with id: " + id + " has been removed.")
	w.Write(bm)

}

func GetRepoInstance(collection string) Repository {
	session := GetSession()
	context = &Context{session}
	col := context.Session.DB(Config.Database).C(collection)
	repo := Repository{col}
	return repo
}

// Handler used to retrieve all messages of a specific user
func GetMessagesByReceiverIdHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("GetMessagesByReceiverIdHandler")
	values := r.URL.Query()
	ids := values["receiver_id"]
	id := ids[0]
	log.Println("Id:", id)

	repo := GetRepoInstance(messageCollection)
	messages := repo.GetMessages(bson.ObjectIdHex(id))
	w.Header().Add("Content-Type", "application/json")

	if len(messages) < 1 {
		w.WriteHeader(http.StatusNoContent)
		bm := composeJsonResponseForNonStructTypes("User has no messages at the moment.")
		w.Write(bm)
		return
	}

	j, err := json.MarshalIndent(&messages, "", "\t")

	if err != nil {
		log.Println("Error marshaling messages:", err)
		w.WriteHeader(http.StatusNoContent)
		bm := composeJsonResponseForNonStructTypes("User has no messages at the moment.")
		w.Write(bm)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(j)

}

func GetMessageByIdHandler(w http.ResponseWriter, r *http.Request) {

}

// Handler used to save and messages to a specific user
func SaveMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("SaveMessageHandler")
	msg := Message{}
	msg.Id = bson.NewObjectId()
	msg.Date = time.Now()
	err := json.NewDecoder(r.Body).Decode(&msg)

	if err != nil {
		log.Println("Error in decoding request body:", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		bm := composeJsonResponseForNonStructTypes(err.Error())
		w.Write(bm)
		return
	}

	repo := GetRepoInstance(messageCollection)
	defer context.Session.Close()
	err = repo.SaveMessage(&msg)

	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		log.Println("Error in saving message:", err)
		w.WriteHeader(http.StatusNoContent)
		bm := composeJsonResponseForNonStructTypes(err.Error())
		w.Write(bm)
		return
	}

	w.WriteHeader(http.StatusCreated)
	bm := composeJsonResponseForNonStructTypes("Message saved!")
	w.Write(bm)
}

func composeJsonResponseForNonStructTypes(msg string) []byte {
	j, _ := json.Marshal(&msg)
	return j
}

func GenerateJWT(user *User) string {
	t := jwt.New(jwt.SigningMethodRS256)
	t.Claims["user"] = user
	t.Claims["exp"] = time.Now().Add(time.Minute * 60).Unix()
	token, err := t.SignedString(SignKey)

	if err != nil {
		log.Println("Error generating jwt:", err)
		return ""
	}
	return token
}
