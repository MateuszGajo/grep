package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type WebData struct {
	Regex   string
	Text    string
	Matches []string
	NFAJson template.JS
}

type NFAData struct {
	InitState string               `json:"initState"`
	States    map[string]StateData `json:"states"`
}

type StateData struct {
	Name        string           `json:"name"`
	Transitions []TransitionData `json:"transitions"`
	IsFinal     bool             `json:"isFinal"`
}

type TransitionData struct {
	To        string `json:"to"`
	Label     string `json:"label"`
	IsEpsilon bool   `json:"isEpsilon"`
}

func webServer() {
	fmt.Println("webserver??")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			regex := r.FormValue("regex")
			text := r.FormValue("text")
			matches, nfa, err := matchLineDetails([]byte(text), regex)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%+v\n", nfa)
			nfaData := convertNFAToData(nfa)
			nfaJson, _ := json.Marshal(nfaData)
			data := WebData{
				Regex:   regex,
				Text:    text,
				Matches: bytesToStrings(matches),
				NFAJson: template.JS(nfaJson),
			}
			tmpl := template.Must(template.ParseFiles("index.html"))

			err = tmpl.Execute(w, data)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("else?")
			http.ServeFile(w, r, "index.html")
		}
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func convertNFAToData(nfa NFA) NFAData {
	statesData := make(map[string]StateData)

	for name, state := range nfa.states {
		transitions := make([]TransitionData, len(state.transitions))
		for i, trans := range state.transitions {
			label := getMatcherLabel(trans.matcher)
			transitions[i] = TransitionData{
				To:        trans.to,
				Label:     label,
				IsEpsilon: trans.isEpsilon,
			}
		}

		statesData[name] = StateData{
			Name:        state.name,
			Transitions: transitions,
			IsFinal:     state.isFinal,
		}
	}

	return NFAData{
		InitState: nfa.initState,
		States:    statesData,
	}
}

func getMatcherLabel(matcher Matcher) string {
	switch m := matcher.(type) {
	case LiteralMatcher:
		return string(m.char)
	case DigitMatcher:
		return "\\d"
	default:
		return "?"
	}
}
