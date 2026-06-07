package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func GetDictionary(client *http.Client, word string) (string, string) {
	baseUrl := os.Getenv("DICTIONARY_FREE_IPA_BASE_URL")
	url := fmt.Sprintf("%s%s", baseUrl, word)

	res, err := client.Get(url)
	if err != nil {
		log.Fatal("Failed to get dictionary: ", err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return word, ""
	}

	var apiResponse []struct {
		Phonetics []struct {
			Text  string `json:"text"`
			Audio string `json:"audio"`
		} `json:"phonetics"`
	}

	err = json.NewDecoder(res.Body).Decode(&apiResponse)
	if err != nil {
		log.Fatal("Failed to decode dictionary: ", err)
	}

	phonetics := apiResponse[0].Phonetics
	// Get the one with audio and phonetic text
	for _, phonetic := range apiResponse[0].Phonetics {
		if phonetic.Text != "" && phonetic.Audio != "" {
			auidoPath := downloadAudio(word, phonetic.Audio)
			fmt.Printf("[DICT] Phonetic: %s, Audio Path: %s\n", phonetic.Text, auidoPath)
			return phonetic.Text, auidoPath
		}
	}

	// Fallback to the first one
	fmt.Printf("[DICT] Phonetic: %s, Audio Path: %s\n", phonetics[0].Text, "")
	return phonetics[0].Text, ""
}

func downloadAudio(word, url string) string {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return "" // caller handles fallback
	}
	defer resp.Body.Close()

	fmt.Printf("[DICT] Audio URL: %s\n", url)

	os.MkdirAll("audio", 0755)
	filename := word + ".mp3"
	filePath := "audio/" + filename
	f, _ := os.Create(filePath)
	defer f.Close()
	io.Copy(f, resp.Body)

	return filePath
}
