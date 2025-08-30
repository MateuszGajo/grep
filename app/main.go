package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

func bytesToStrings(bs [][]byte) []string {
	var ss []string
	for _, b := range bs {
		ss = append(ss, string(b))
	}
	return ss
}

func main() {
	web := false
	if os.Args[1] == "-web" {
		web = true
	}

	if web {
		webServer()
	} else {
		cli()
	}
}

func cli() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2)
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	matches, _, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if len(matches) == 0 {
		os.Exit(1)
	}

}

func matchLine(line []byte, pattern string) ([][]byte, NFA, error) {
	parser := Parser{conversion: Conversion{}, pattern: pattern, pos: 0}
	nfa, err := parser.parse()

	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	isStartAnchor := false
	if len(pattern) > 0 && pattern[0] == '^' {
		isStartAnchor = true
	}

	matches := nfa.findAllMatches(line, isStartAnchor)

	return matches, nfa, err
}

type NFATransition struct {
	to      string
	matcher Matcher
}

type NFA struct {
	initState string
	endStates []string
	states    map[string]State
}

type State struct {
	name        string
	transitions []NFATransition
	isFinal     bool
}

var stateCounter int = 0

func NewState() State {
	stateCounter++
	return State{
		name:        fmt.Sprintf("q%d", stateCounter),
		transitions: []NFATransition{},
	}
}

func (n *NFA) addStates(states []State) {
	for _, state := range states {
		n.states[state.name] = state
	}
}

func (n *NFA) addTransition(from State, to State, matcher Matcher) error {

	fromState, ok := n.states[from.name]
	if !ok {
		return fmt.Errorf("could find state from")
	}

	fromState.transitions = append(fromState.transitions, NFATransition{to: to.name, matcher: matcher})

	n.states[from.name] = fromState

	return nil
}

func (n *NFA) addTransitionPrio(from State, to State, matcher Matcher) error {

	fromState, ok := n.states[from.name]
	if !ok {
		return fmt.Errorf("could find state from")
	}

	fromState.transitions = append([]NFATransition{{to: to.name, matcher: matcher}}, fromState.transitions...)

	n.states[from.name] = fromState

	return nil
}

func (n *NFA) setInitState(state State) {
	n.initState = state.name
}

func (n *NFA) setFinalStates(states []State) error {
	n.endStates = []string{}
	for _, state := range states {
		state, ok := n.states[state.name]

		if !ok {
			return fmt.Errorf("state doesnt exist")
		}

		state.isFinal = true

		n.states[state.name] = state
		n.endStates = append(n.endStates, state.name)
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

func (n *NFA) run(line []byte, index int) (bool, []byte, int) {
	stack := Stack{}
	stack.push(n.states[n.initState], 0)
	line = line[index:]
	for stack.length() > 0 {
		item := stack.pop()
		if item.currentState.isFinal {
			return true, line[:item.i], index + item.i
		}

		for i := len(item.currentState.transitions) - 1; i >= 0; i-- {
			transition := item.currentState.transitions[i]
			if (item.i < len(line) || transition.matcher.isEpsilon()) && transition.matcher.match(line, item.i) {
				newIndex := item.i
				if !transition.matcher.isEpsilon() {
					newIndex += 1
				}
				toState := n.states[transition.to]
				stack.push(toState, newIndex)
			}
		}
	}
	return false, nil, 0
}

func (n *NFA) findAllMatches(input []byte, isStartAnchor bool) [][]byte {
	matches := [][]byte{}
	counter := 0
	i := 0
	for i < len(input) {
		counter++
		if counter > 20 {
			break
		}
		if i > 0 && isStartAnchor {
			continue
		}
		ok, match, index := n.run(input, i)

		if ok {
			matches = append(matches, match)
			i = index
		} else {
			i++
		}

	}

	return matches
}

func (n *NFA) appendNfa(nfa NFA, unionStateName string) {
	for _, item := range nfa.states {
		if item.name == nfa.initState {
			continue
		}
		n.states[item.name] = item
	}

	for _, transition := range nfa.states[nfa.initState].transitions {
		n.addTransition(n.states[unionStateName], n.states[transition.to], transition.matcher)
	}

	newEndStates := []string{}

	for _, item := range n.endStates {

		if item == unionStateName {
			newEndStates = append(newEndStates, nfa.endStates...)
			unionState := n.states[unionStateName]
			unionState.isFinal = false
			n.states[unionStateName] = unionState
		} else {
			newEndStates = append(newEndStates, item)
		}
	}
	n.endStates = newEndStates

}

// ------------------ Parser ------------------
// Builds an AST based on the following regex grammar:
//
// operator precedence:
//
//   Alternation     → Concatenation ("|" Concatenation)*
//   Concatenation   → QuantifiedAtom+
//   QuantifiedAtom  → Atom Quantifier?
//   Atom            → Literal | Group | CharClass | Escape
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
//		- Escape	   : an espcped characters such as \d \w
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

func (c Conversion) oneStepNFA(matcher Matcher) (NFA, error) {
	nfa := NFA{states: map[string]State{}}
	a := NewState()
	b := NewState()
	nfa.addStates([]State{a, b})
	nfa.setInitState(a)
	err := nfa.setFinalStates([]State{b})
	if err != nil {
		return nfa, err
	}
	err = nfa.addTransition(a, b, matcher)

	return nfa, err
}

type Parser struct {
	pattern    string
	pos        int
	conversion Conversion
}

func (p Parser) isEnd() bool {
	return p.pos >= len(p.pattern)
}

func (p Parser) peekNext() byte {
	index := p.pos
	index++
	return p.pattern[index]
}

func (p Parser) isNextEnd() bool {
	index := p.pos
	index++
	return index >= len(p.pattern)
}

func (p *Parser) parse() (NFA, error) {
	stateCounter = 0
	left, err := p.parseRepeat()
	if err != nil {
		return NFA{}, err
	}
	for !p.isEnd() {
		right, err := p.parseRepeat()
		if err != nil {
			return NFA{}, err
		}
		left.appendNfa(right, left.endStates[0])

	}

	return left, nil
}

func (p *Parser) parseRepeat() (NFA, error) {
	leftAtom, err := p.parseAtom()

	if p.isEnd() {
		return leftAtom, err
	}

	c := p.pattern[p.pos]

	switch c {
	case '+':
		/*
					  ┌────────ε────────┐
					  ▼          	    │
			(q1) -ε> (q2) -condition-> (q3) -ε> ((q4))

			1. Create start state q1
			2. Create end state q4
			3. Add epsilon transition from q1 to q2
			4. Add epsilon transition from q3 to q2 (loop repetition), by default it's greedy so prioritize it rather than exit the loop
			5. Add epsilon transition from q3 to q4

		*/

		q1 := NewState()
		q4 := NewState()
		q4.isFinal = true

		leftAtom.addStates([]State{q1, q4})
		currentInitState := leftAtom.initState
		leftAtom.setInitState(q1)
		leftAtom.addTransition(q1, leftAtom.states[currentInitState], EpsilonMatcher{})
		// Greedy matcher, need to be added first
		leftAtom.addTransition(leftAtom.states[leftAtom.endStates[0]], leftAtom.states[currentInitState], EpsilonMatcher{})
		leftAtom.addTransition(leftAtom.states[leftAtom.endStates[0]], q4, EpsilonMatcher{})

		last := leftAtom.states[leftAtom.endStates[0]]
		last.isFinal = false
		leftAtom.states[leftAtom.endStates[0]] = last

		leftAtom.setFinalStates([]State{q4})
		p.pos++
	case '?':
		/*

			(q1) -ε> (q2) -condition-> (q3) -ε> ((q4))
			  │                                    ▲
			  └────────────────ε───────────────────┘

			1. Create start state q1
			2. Create end state q4
			3. Add epsilon transition from q1 to q2
			4. Add epsilon transition from q1 to q4 (exit), by default it's greedy so prioritize entering condition
			5. Add epsilon transition from q3 to q4

		*/

		q1 := NewState()
		q4 := NewState()
		q4.isFinal = true

		leftAtom.addStates([]State{q1, q4})
		currentInitState := leftAtom.initState
		leftAtom.setInitState(q1)
		leftAtom.addTransition(q1, leftAtom.states[currentInitState], EpsilonMatcher{})
		leftAtom.addTransition(q1, q4, EpsilonMatcher{})
		// Greedy matcher, need to be added first
		leftAtom.addTransition(leftAtom.states[leftAtom.endStates[0]], q4, EpsilonMatcher{})

		last := leftAtom.states[leftAtom.endStates[0]]
		last.isFinal = false
		leftAtom.states[leftAtom.endStates[0]] = last

		leftAtom.setFinalStates([]State{q4})
		p.pos++
	}

	return leftAtom, err
}

func (p *Parser) parseAtom() (NFA, error) {
	switch p.pattern[p.pos] {
	case '\\':
		return p.parseEscape()
	case '[':
		return p.parseCharGroup()
	case '^':
		return p.parseCaretAnchor()
	case '$':
		return p.parseDollarAnchor()
	case '.':
		return p.parseDot()
	default:
		return p.parseLiteral()
	}
}

func (p *Parser) parseDot() (NFA, error) {
	matcher := AnyCharMatcher{}
	p.pos++
	return p.conversion.oneStepNFA(matcher)
}

func (p *Parser) parseEscape() (NFA, error) {
	p.pos++
	esc := p.pattern[p.pos]
	p.pos++
	switch esc {
	case 'd':
		return p.conversion.oneStepNFA(NewDigitMatcher())
	case 'w':
		return p.conversion.oneStepNFA(NewWordMatcher())
	default:
		return p.parseLiteral()
	}
}

func (p *Parser) parseCharGroup() (NFA, error) {
	if !strings.Contains(p.pattern[p.pos:], "]") {
		return NFA{}, fmt.Errorf("missing ] for char group")
	}
	start := p.pos
	p.pos++
	isNegative := false
	if p.pattern[p.pos] == '^' {
		isNegative = true
		p.pos++
	}
	ranges := []CharRange{}
	chars := []byte{}
	for !p.isEnd() && p.pattern[p.pos] != ']' {
		char := p.pattern[p.pos]
		switch char {
		case '\\':
			p.pos++
			if p.isEnd() {
				return NFA{}, fmt.Errorf("end condition after //")
			}
			esc := p.pattern[p.pos]
			switch esc {
			case 'd':
				ranges = append(ranges, digitMatcherRanges...)
			case 'w':
				ranges = append(ranges, wordMarcherRanges...)
				chars = append(chars, wordMatcherChars...)
			}
		default:
			if !p.isNextEnd() && p.peekNext() == '-' {
				p.pos++
				if p.isNextEnd() {
					return NFA{}, fmt.Errorf("end condition of range its the last char in pattern")
				}
				p.pos++
				nextChar := p.pattern[p.pos]
				ranges = append(ranges, CharRange{from: char, to: nextChar})
			} else {
				chars = append(chars, char)
			}

		}
		p.pos++
	}
	p.pos++
	charGroupMatcher := NewCharacterGroupMatcher(ranges, chars, isNegative, p.pattern[start:p.pos])

	return p.conversion.oneStepNFA(charGroupMatcher)
}

func (p *Parser) parseLiteral() (NFA, error) {
	matcher := LiteralMatcher{char: p.pattern[p.pos]}
	p.pos++
	return p.conversion.oneStepNFA(matcher)
}

func (p *Parser) parseDollarAnchor() (NFA, error) {
	matcher := EndOfStringMatcher{}
	p.pos++
	return p.conversion.oneStepNFA(matcher)
}

func (p *Parser) parseCaretAnchor() (NFA, error) {
	matcher := StartOfStringMatcher{}
	p.pos++
	return p.conversion.oneStepNFA(matcher)
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
	match(b []byte, index int) bool
	isEpsilon() bool
}

type EpsilonMatcher struct {
	char byte
}

func (lm EpsilonMatcher) match(b []byte, index int) bool {
	return true
}

func (lm EpsilonMatcher) isEpsilon() bool {
	return true
}

type LiteralMatcher struct {
	char byte
}

func (lm LiteralMatcher) match(b []byte, index int) bool {
	return lm.char == b[index]
}

func (lm LiteralMatcher) isEpsilon() bool {
	return false
}

type DigitMatcher struct {
	ranges []CharRange
}

func (dm DigitMatcher) match(b []byte, index int) bool {
	return matchRanges(dm.ranges, b[index])
}

func (dm DigitMatcher) isEpsilon() bool {
	return false
}

func NewDigitMatcher() WordMatcher {
	return WordMatcher{
		ranges: digitMatcherRanges,
	}
}

type WordMatcher struct {
	ranges []CharRange
	chars  []byte
}

func (wm WordMatcher) match(b []byte, index int) bool {
	return matchRanges(wm.ranges, b[index]) || matchChars(wm.chars, b[index])
}

func (wm WordMatcher) isEpsilon() bool {
	return false
}

var (
	wordMarcherRanges  = []CharRange{{from: '0', to: '9'}, {from: 'a', to: 'z'}, {from: 'A', to: 'Z'}}
	wordMatcherChars   = []byte{'_'}
	digitMatcherRanges = []CharRange{{from: '0', to: '9'}}
)

func NewWordMatcher() WordMatcher {
	return WordMatcher{
		ranges: wordMarcherRanges,
		chars:  wordMatcherChars,
	}
}

type CharRange struct {
	from byte
	to   byte
}

func (charRange CharRange) match(b byte) bool {
	if b >= charRange.from && b <= charRange.to {
		return true
	}
	return false
}

type CharacterGroupMatcher struct {
	ranges     []CharRange
	chars      []byte
	isNegative bool
	label      string
}

func (cgm CharacterGroupMatcher) match(b []byte, index int) bool {
	base := matchChars(cgm.chars, b[index]) || matchRanges(cgm.ranges, b[index])

	if cgm.isNegative {
		return !base
	}

	return base
}

func (lm CharacterGroupMatcher) isEpsilon() bool {
	return false
}

func NewCharacterGroupMatcher(ranges []CharRange, chars []byte, isNegative bool, label string) CharacterGroupMatcher {
	return CharacterGroupMatcher{
		ranges:     ranges,
		chars:      chars,
		isNegative: isNegative,
		label:      label,
	}
}

func matchRanges(ranges []CharRange, b byte) bool {
	for _, item := range ranges {
		if item.match(b) {
			return true
		}
	}
	return false
}

func matchChars(chars []byte, b byte) bool {
	return slices.Contains(chars, b)
}

type StartOfStringMatcher struct{}

func (startOfStringMatcher StartOfStringMatcher) match(b []byte, index int) bool {
	return index == 0
}

func (lm StartOfStringMatcher) isEpsilon() bool {
	return true
}

type EndOfStringMatcher struct{}

func (endOfStringMatcher EndOfStringMatcher) match(b []byte, index int) bool {

	return len(b) == index
}

func (lm EndOfStringMatcher) isEpsilon() bool {
	return true
}

type AnyCharMatcher struct{}

func (anyCharMatcher AnyCharMatcher) match(b []byte, index int) bool {
	return true
}

func (anyCharMatcher AnyCharMatcher) isEpsilon() bool {
	return false
}
