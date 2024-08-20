package sheets

import (
	"encoding/json"
	"log"
	"os"

	"golang.org/x/oauth2"
)

var tokenPath = "token.json"

// TODO token from keystring?
func RetrieveTokenFromFile() (*oauth2.Token, error) {

	tokenFile, err := os.Open(tokenPath)

	if err != nil {
		return nil, err
	}

	defer tokenFile.Close()

	tok := &oauth2.Token{}

	err = json.NewDecoder(tokenFile).Decode(tok)

	return tok, err

}

func saveToken(token *oauth2.Token) {

	f, err := os.OpenFile(tokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}

	defer f.Close()

	json.NewEncoder(f).Encode(token)
}
