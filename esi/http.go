package esi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sethgrid/pester"
)

func fetchURL(client *pester.Client, url string, r interface{}) error {
	log.Printf("Fetching %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", "go-evepraisal")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error talking to esi: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(r)
	defer resp.Body.Close()
	return err
}
