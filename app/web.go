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
			matches, nfa, err := matchLine([]byte(text), regex)
			if err != nil {
				log.Fatal(err)
			}

			nfaData := convertNFAToData(nfa)
			
			// Debug: Print NFA structure to terminal
			fmt.Printf("=== NFA DEBUG (regex: %s) ===\n", regex)
			fmt.Printf("Init State: %s\n", nfaData.InitState)
			fmt.Println("ALL STATES:")
			for stateName, state := range nfaData.States {
				fmt.Printf("State %s: isFinal=%t, transitions=%d\n", stateName, state.IsFinal, len(state.Transitions))
				for _, t := range state.Transitions {
					fmt.Printf("  -> %s (label: \"%s\", epsilon: %t)\n", t.To, t.Label, t.IsEpsilon)
				}
			}
			fmt.Println("===================")
			
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
				IsEpsilon: trans.matcher.isEpsilon(),
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
	if matcher.isEpsilon() {
		return "ε"
	}
	switch m := matcher.(type) {
	case LiteralMatcher:
		return string(m.char)
	case DigitMatcher:
		return "\\d"
	case WordMatcher:
		return "\\w"
	case CharacterGroupMatcher:
		return m.label
	default:
		return "?"
	}
}
