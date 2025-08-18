package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Ensures gofmt doesn't remove the "bytes" import above (feel free to remove this!)
var _ = bytes.ContainsAny

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}

	// default exit code is 0 which means success
}

func matchLine(line []byte, pattern string) (bool, error) {
	parser := Parser{conversion: Conversion{}, pattern: pattern, pos: 0}
	nfa := parser.parse()
	var ok bool

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	// Uncomment this to pass the first stage
	ok = nfa.run(line)

	return ok, nil
}

type NFATransition struct {
	to        string
	matcher   Matcher
	isEpsilon bool
}

type NFA struct {
	initState string
	states    map[string]State
}

type State struct {
	name        string
	transitions []NFATransition
	isFinal     bool
}

var stateCounter int = 0

func NewState() *State {
	stateCounter++
	return &State{
		name:        fmt.Sprintf("q%d", stateCounter),
		transitions: []NFATransition{},
	}
}

func (n *NFA) addStates(states []State) {
	for _, state := range states {
		n.states[state.name] = state
	}
}

func (n *NFA) addTransition(from State, to State, matcher Matcher) {

	fromState, ok := n.states[from.name]
	if !ok {
		panic("could find state from")
	}

	fromState.transitions = append(fromState.transitions, NFATransition{to: to.name, matcher: matcher})

	n.states[from.name] = fromState
}

func (n *NFA) setInitState(state State) {
	n.initState = state.name
}

func (n *NFA) setFinalStates(states []State) error {
	for _, state := range states {
		state, ok := n.states[state.name]

		if !ok {
			return fmt.Errorf("state doesnt exist")
		}

		state.isFinal = true

		n.states[state.name] = state
	}

	return nil
}

type StackData struct {
	currentState State
	i            int
}

type Stack struct {
	data []StackData
}

func (n *NFA) run(line []byte) bool {
	stack := Stack{}
	stack.push(n.states[n.initState], 0)

	for stack.length() > 0 {
		item := stack.pop()
		if item.currentState.isFinal {
			return true
		}

		for _, transition := range item.currentState.transitions {

			if transition.matcher.match(line[item.i]) {
				newIndex := item.i
				if !transition.isEpsilon {
					newIndex += 1
				}
				toState := n.states[transition.to]
				stack.push(toState, newIndex)
			}
		}
	}
	return false
}

// ------------------ Parser ------------------
// Builds an AST based on the following regex grammar:
//
// operator precedence:
//
//   Alternation     → Concatenation ("|" Concatenation)*
//   Concatenation   → QuantifiedAtom+
//   QuantifiedAtom  → Atom Quantifier?
//   Atom            → Literal | Group | CharClass
//   Quantifier      → '?' | '*' | '+' | '{' Number (',' Number?)? '}'
//
// ---------------------------------------------------------
// Rule Explanations:
//
// 1. Atom
//    The most basic building block of regex:
//      - Literal      : a single character (e.g. "a", "9", ".")
//      - Group        : a sub-expression in parentheses (e.g. "(ab|cd)")
//      - CharClass    : a character set in brackets (e.g. "[a-z0-9]", "[^abc]")
//
// Focus on the next steps later; for now, only Atom is relevant
// 2. Quantifier
//    Specifies repetition of the preceding Atom:
//      - '?'          : 0 or 1 occurrence
//      - '*'          : 0 or more occurrences
//      - '+'          : 1 or more occurrences
//      - '{n}', '{n,m}': exact or ranged number of repetitions
//
// 3. QuantifiedAtom
//    Combines an Atom with an optional Quantifier.
//    Example:
//      - 'a*'      : Atom 'a' with '*' quantifier
//      - '(dog)+'  : Atom is the group '(dog)' with '+' quantifier
//      - '[0-9]?'  : Atom is character class '[0-9]' with '?' quantifier
//
// 4. Concatenation
//    Matches a sequence of QuantifiedAtoms in order.
//    Example:
//      - 'abc'    : 'a' then 'b' then 'c'
//      - 'a\d+'   : 'a' then '\d+' (digit repeated 1 or more)
//
// 5. Alternation
//    Provides a choice between multiple Concatenations using '|'.
//    Example:
//      - 'cat|dog' : matches 'cat' OR 'dog'
//      - 'a(b|c)d' : matches 'abd' OR 'acd'
//
// =========================================================

type Conversion struct {
}

func (c Conversion) oneStepNFA(matcher Matcher) NFA {
	nfa := NFA{states: map[string]State{}}
	a := *NewState()
	b := *NewState()
	nfa.addStates([]State{a, b})
	nfa.setInitState(a)
	nfa.setFinalStates([]State{b})
	nfa.addTransition(a, b, matcher)

	return nfa
}

type Parser struct {
	pattern    string
	pos        int
	conversion Conversion
}

func (p Parser) parse() NFA {
	return p.parseAtom()
}

func (p Parser) parseAtom() NFA {
	switch p.pattern[p.pos] {
	default:
		return p.parseLiteral(p.pattern[p.pos])
	}
}

func (p Parser) parseLiteral(b byte) NFA {
	return p.conversion.oneStepNFA(LiteralMatcher{char: b})
}

// ------------------ Stack ------------------

func (s *Stack) push(state State, i int) {
	s.data = append(s.data, StackData{currentState: state, i: i})
}

func (s *Stack) pop() StackData {
	lastItem := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]

	return lastItem
}

func (s Stack) length() int {
	return len(s.data)
}

// ------------------ Matcher ------------------

type Matcher interface {
	match(b byte) bool
}

type LiteralMatcher struct {
	char byte
}

func (lm LiteralMatcher) match(b byte) bool {
	return lm.char == b
}
