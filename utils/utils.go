package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

func ParseBody(body io.Reader, out interface{}) error {
	if err := json.NewDecoder(body).Decode(out); err != nil {
		return err
	}
	return nil
}

// go to json - sent response
func EncodeJSONBody(w http.ResponseWriter, data interface{}) error {
	return json.NewEncoder(w).Encode(data)
}

type Error struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func RespondError(w http.ResponseWriter, err error, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	var errStr string

	if err != nil {
		errStr = err.Error()
	}

	NewError := Error{
		Error:      errStr,
		Message:    message,
		StatusCode: statusCode,
	}

	if err := EncodeJSONBody(w, NewError); err != nil {
		fmt.Printf("error is %v", err)
	}
}

func RespondJSON(w http.ResponseWriter, StatusCode int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(StatusCode)
	if body != nil {
		if err := json.NewEncoder(w).Encode(body); err != nil {
			fmt.Printf("error is %v", err)
		}
	}
}

func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword), []byte(password))
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	return string(hash), err
}

func GenerateJWT(sessionID, userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"role":       role,
		"exp":        time.Now().Add(time.Minute * 60).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}

var validate = validator.New()

func ValidateStruct(body interface{}) error {
	return validate.Struct(body)
}
