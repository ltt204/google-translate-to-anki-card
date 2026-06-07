package main

const ankiConnectURL = "http://localhost:8765"
const modelName = "VocabPro"

var wantedFields = []string{"Word", "IPA", "Part_of_Speech", "Definition_VI", "Sentence_Front", "Synonym_Hint", "Structure_Hint", "Audio"}

const cardFront = `<div class="sentence">{{Sentence_Front}}</div>
<div class="hint">[ {{Part_of_Speech}} | {{Synonym_Hint}}{{#Structure_Hint}} | {{Structure_Hint}}{{/Structure_Hint}} ]</div>
{{Audio}}`

const cardBack = `{{FrontSide}}<hr>
<div class="word">{{Word}}</div>
<div class="ipa">{{IPA}}</div>
<div class="pos">{{Part_of_Speech}}</div>
<div class="def">{{Definition_VI}}</div>`

const cardCSS = `
.card { font-family: Arial; font-size: 18px; text-align: center; color: #222; }
.sentence { font-size: 22px; margin-bottom: 12px; }
.hint { font-size: 14px; color: #888; font-style: italic; margin-bottom: 12px; }
hr { border-color: #ddd; margin: 16px 0; }
.word { font-size: 28px; font-weight: bold; }
.ipa  { color: #888; font-size: 15px; margin: 4px 0; }
.pos  { color: #aaa; font-style: italic; font-size: 13px; }
.def  { color: #4CAF50; margin-top: 10px; font-size: 16px; }
`
