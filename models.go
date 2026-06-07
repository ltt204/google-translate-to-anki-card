package main

import "encoding/json"

type GoogleTranslateCSVRecord struct {
	OriginalLang   string `json:"original_lang"`
	TargetLang     string `json:"target_lang"`
	OriginWord     string `json:"origin_word"`
	TranslatedWord string `json:"translated_word"`
}

type AnkiRecord struct {
	Word          string `json:"word"`
	Phonetic      string `json:"phonetic"`
	PartOfSpeech  string `json:"part_of_speech"`
	Definition    string `json:"definition"`
	SentenceFront string `json:"sentence_front"`
	SynonymHint   string `json:"synonym_hint"`
	StructureHint string `json:"structure_hint"`
	AudioFilePath string `json:"audio_file_path"`
	AudioFileName string `json:"audio_file_name"`
	Deck          string `json:"deck"`
	Tags          string `json:"tags"`
}

type DictionaryAPIResponse struct {
	Phonetic      string `json:"phonetic"`
	AudioFilePath string `json:"audio_file_path"`
}

type ExampleResult struct {
	PartOfSpeech  string `json:"part_of_speech"`
	Sentence      string `json:"sentence"`
	SynonymHint   string `json:"synonym_hint"`
	StructureHint string `json:"structure_hint"`
}

type LLMResponse struct {
	ExampleResults []ExampleResult `json:"results"`
}

// Anki
type ankiRequest struct {
	Action  string `json:"action"`
	Version int    `json:"version"`
	Params  any    `json:"params,omitempty"`
}

type ankiResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *string         `json:"error"`
}
