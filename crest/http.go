package crest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func fetchURL(client *http.Client, url string, r interface{}) error {
	log.Printf("Fetching %s", url)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error talking to crest: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(r)
	defer resp.Body.Close()
	return err
}
