package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"google.golang.org/genai"
)

// This created for only work with Google Transalte's csv format.
// You can get the csv by exporting the saved words to csv from Google Transalte's saved words section.
// Sample format from Google Translate:
//   The abstaction (not the HEADER): original_lang,origin_word,target_lang,translated_word
//   English,rational,Vietnamese,hợp lý --> 1 row, also the first row of the csv file
//   English,ambiguous,Vietnamese,mơ hồ
//   English,resilient,Vietnamese,kiên cường

// The target formats, i am persionally aiming for, is:
//   word,ipa,part_of_speech,definition_vi,sentence_front,audio_file,deck,tags
//   rational,/ˈræʃənl/,adjective,Hợp lý - dựa trên lý trí,We need to make a [......] decision based on data.,rational.mp3,EN→VI,en-vi adjective
//   ambiguous,/æmˈbɪɡjuəs/,adjective,Mơ hồ - không rõ ràng,The instructions were [......] so nobody knew what to do.,ambiguous.mp3,EN→VI,en-vi adjective
//   resilient,/rɪˈzɪliənt/,adjective,Kiên cường - có khả năng phục hồi,She proved to be [......] after the setback.,resilient.mp3,EN→VI,en-vi adjective

func TransformToAnkiFormat(ctx context.Context, client *http.Client, chat *genai.Chat, filename string, knownWords map[string]bool) []AnkiRecord {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", filename, err)
	}

	ch := make(chan GoogleTranslateCSVRecord, 10)
	results := make(chan AnkiRecord, 10)

	// Goroutine 1: Feed the channel with the raw data from the Google Translate csv file
	go func() {
		reader := csv.NewReader(file)
		for {
			record, err := reader.Read()

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalf("Failed to read record: %v", err)
			}

			word := strings.ToLower(record[2])
			if knownWords[word] {
				fmt.Printf("[SKIP] %s already in deck\n", record[2])
				continue
			}

			googleTranslateRecord := GoogleTranslateCSVRecord{
				OriginalLang:   record[0],
				TargetLang:     record[1],
				OriginWord:     record[2],
				TranslatedWord: record[3],
			}
			ch <- googleTranslateRecord

		}
		close(ch)
	}()

	// Goroutine 2: Take the data from the channel and transform it to Anki format
	var wg sync.WaitGroup
	go func() {
		for raw := range ch {
			wg.Add(1)
			go func(raw GoogleTranslateCSVRecord) {
				defer wg.Done()

				var dictPhonetic, dictAudioPath, dictAudioFileName string
				var llmsResponse []ExampleResult

				var inner sync.WaitGroup
				inner.Add(2)
				go func() {
					defer inner.Done()
					dictPhonetic, dictAudioPath = GetDictionary(client, raw.OriginWord)
					dictAudioFileName = raw.OriginWord + ".mp3"
				}()
				go func() {
					defer inner.Done()
					llmsResponse = GetExamples(ctx, chat, raw.OriginWord, raw.TranslatedWord).ExampleResults

				}()
				inner.Wait()
				// Format the data to Anki format
				for _, entry := range llmsResponse {
					results <- AnkiRecord{
						NoteID:        raw.OriginWord + "_" + entry.PartOfSpeech,
						Word:          raw.OriginWord,
						Definition:    raw.TranslatedWord,
						PartOfSpeech:  entry.PartOfSpeech,
						SentenceFront: entry.Sentence,
						SynonymHint:   entry.SynonymHint,
						StructureHint: entry.StructureHint,
						Phonetic:      dictPhonetic,
						AudioFilePath: dictAudioPath,
						AudioFileName: dictAudioFileName,
						Deck:          "vocabulary",
						Tags:          "en-vi " + entry.PartOfSpeech,
					}
				}
			}(raw)
		}
		wg.Wait()
		close(results)
	}()

	var ankiRecords []AnkiRecord
	for r := range results {
		ankiRecords = append(ankiRecords, r)
	}
	return ankiRecords
}

func WriteCSV(records []AnkiRecord) {
	file, err := os.Create("anki.csv")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range records {
		record := []string{
			record.Word,
			record.Phonetic,
			record.PartOfSpeech,
			record.Definition,
			record.SentenceFront,
			record.AudioFilePath,
			record.AudioFileName,
			record.Deck,
			record.Tags,
		}
		if err := writer.Write(record); err != nil {
			log.Fatalf("Error writing record to csv: %v", err)
		}
	}
}
