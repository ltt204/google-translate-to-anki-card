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

	// Get the one with audio and phonetic text
	for _, phonetic := range apiResponse[0].Phonetics {
		if phonetic.Text != "" && phonetic.Audio != "" {
			audioPath := downloadAudio(word, phonetic.Audio)
			fmt.Printf("[DICT] Phonetic: %s, Audio Path: %s\n", phonetic.Text, audioPath)
			return phonetic.Text, audioPath
		}
	}

	// Fallback: return first phonetic text with no audio
	for _, phonetic := range apiResponse[0].Phonetics {
		if phonetic.Text != "" {
			fmt.Printf("[DICT] Phonetic: %s, Audio Path: \n", phonetic.Text)
			return phonetic.Text, ""
		}
	}

	// No phonetics at all
	fmt.Printf("[DICT] No phonetics found for %s\n", word)
	return "", ""
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
