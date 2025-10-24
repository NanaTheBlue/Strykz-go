package authapi

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/nanagoboiler/models"
)

func validateUsername(username string) error {
	var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,15}$`)

	if len(username) > 15 || len(username) < 3 {
		return errors.New("Invalid Username Length")

	}

	if !usernameRegex.MatchString(username) {
		return errors.New("Invalid Characters in Username")
	}

	return nil

}

func validatePassword(password string, confirmpassword string) error {

	// Just basic password validation for now
	if len(password) < 8 {
		return errors.New("password length is to short")

	} else if len(password) > 16 {
		return errors.New("password length is to long")
	}

	if password != confirmpassword {
		return errors.New("passwords dont match")
	}

	return nil
}

func validateEmail(email string) error {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil

}

func validateRegistration(req *models.RegisterRequest) error {

	err := validatePassword(req.Password, req.ConfirmPassword)
	if err != nil {
		return err
	}
	err = validateEmail(req.Email)
	if err != nil {
		return err
	}
	err = validateUsername(req.Username)
	if err != nil {
		return err
	}

	return nil
}

func setCookie(w http.ResponseWriter, name string, value string, expires time.Time, samesite http.SameSite, secure bool, httponly bool) {

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  expires,
		SameSite: samesite,
		HttpOnly: httponly,
		Secure:   secure,
	})

}
