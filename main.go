package main

import (
	"log"
	"net/http"
	"time"

	. "github.com/GermanMontejo/chatapp/handler"
	. "github.com/GermanMontejo/chatapp/utils"
	"github.com/gorilla/mux"
)

func main() {
	m := mux.NewRouter()
	StartUp()
	fs := http.FileServer(http.Dir("public"))

	m.Handle("/public/", fs)
	m.HandleFunc("/users", GetUsersHandler).Methods("GET")
	m.HandleFunc("/users", DeleteUserHandler).Methods("DELETE")
	m.Handle("/users/login", AuthMiddleware(http.HandlerFunc(LoginHandler))).Methods("POST")
	m.HandleFunc("/users/messages", GetMessagesByReceiverIdHandler).Methods("GET")
	m.HandleFunc("/users/messages", SaveMessageHandler).Methods("POST")
	m.HandleFunc("/users/signup", SignupHandler).Methods("POST")
	log.Println("Listening...")

	server := http.Server{
		Addr:        Config.Server,
		ReadTimeout: time.Second * 60,
		Handler:     m,
	}

	log.Println("Server Address:", server.Addr)
	server.ListenAndServe()
}
