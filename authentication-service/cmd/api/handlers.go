package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.ReadJson(w, r, &requestPayload)
	if err != nil {
		app.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	log.Printf("requestPayload: %s", requestPayload)

	// validate user against the DB
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		log.Println(err)
		app.ErrorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	log.Println("Got user email...")

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.ErrorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.WriteJson(w, http.StatusAccepted, payload)
}
