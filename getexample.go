package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"google.golang.org/genai"
)

func GetExamples(ctx context.Context, chat *genai.Chat, word string, translation string) LLMResponse {
	systemPrompt := fmt.Sprintf(`
Word: "%s"
Translation: "%s"

Task: For this word, return cards for the top 1-3 most commonly used parts of speech.

For each entry return:
- part_of_speech: the part of speech (noun, verb, adjective, etc.)
- sentence: a natural B2-level example sentence with the target word replaced by [......]
- synonym_hint: a short English synonym or meaning paraphrase (do NOT use the word itself or its translation)
- structure_hint: an optional note on etymology, metaphor, or domain (e.g. "abstract train metaphor", "from Latin root meaning X", "common in academic writing"). Leave empty string if nothing useful to add.

## Example

Input:
Word: "derail"
Translation: "làm trật bánh"

Output:
{
	"results": [
		{
			"part_of_speech": "verb",
			"sentence": "The system crash threatens to [......] our software launch.",
			"synonym_hint": "ruin a process / cause something to go off track",
			"structure_hint": "abstract train metaphor — literally means a train leaving its tracks"
		},
		{
			"part_of_speech": "noun",
			"sentence": "The sudden [......] of the peace talks shocked diplomats worldwide.",
			"synonym_hint": "a breakdown or disruption of a process",
			"structure_hint": ""
		}
	]
}

## Rules

1. Always include the primary, most common part of speech.
2. Only add alternative parts of speech if they are frequently used in modern professional or academic English.
   - YES: "conduct" as noun AND verb — both common.
   - NO: "anecdote" as verb — extremely rare, skip it.
3. Maximum 3 entries.
4. The synonym_hint must help the learner guess the word without giving it away directly.
5. Return ONLY valid JSON. No explanation, no markdown, no extra text.
`, word, translation)

	rawResponse, err := chat.SendMessage(ctx, genai.Part{
		Text: systemPrompt,
	})
	if err != nil {
		log.Fatal("Failed to generate examples: ", err)
	}

	response := &LLMResponse{}
	err = json.NewDecoder(strings.NewReader(rawResponse.Text())).Decode(response)
	if err != nil {
		log.Fatal("Failed to decode examples: ", err)
	}

	fmt.Printf("[LLM] Response: %+v\n", *response)

	return *response
}
