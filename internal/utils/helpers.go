package utils

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode"
)

func RespondWithJSONHelper(w http.ResponseWriter, statuscode int, payload interface{}) error {
    response, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.WriteHeader(statuscode)
    w.Write(response)
    return nil
}

func RespondWithErrorHelper(w http.ResponseWriter, statuscode int, msg string) error {
    return RespondWithJSONHelper(w, statuscode, map[string]string{"error": msg})
}

func CheckBadChirpLang(bodyString string) string {
    badWords := map[string]bool{
        "kerfuffle" : true,
        "sharbert" : true,
        "fornax" : true,
    }
    var currentWord strings.Builder
    var result strings.Builder

    for _, char := range bodyString {
        if unicode.IsLetter(char) {
            currentWord.WriteRune(char)
        } else {
            if currentWord.Len() > 0 {
                word := strings.ToLower(currentWord.String())
                if badWords[word] {
                    result.WriteString("****")
                } else {
                    result.WriteString(currentWord.String())
                }
                currentWord.Reset()
            }
            result.WriteRune(char)
        }
    }

    if currentWord.Len() > 0 {
        word := strings.ToLower(currentWord.String())
        if badWords[word] {
            result.WriteString("****")
        } else {
            result.WriteString(currentWord.String())
        }
    }

    return result.String()

}
