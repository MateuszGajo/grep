package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
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

func readFile(filename string) ([][]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("problem opening file %v, err:%v", filename, err)
	}
	sc := bufio.NewScanner(file)

	lines := [][]byte{}

	for sc.Scan() {
		lines = append(lines, sc.Bytes())
	}

	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}

	return lines, nil
}

func cli() {
	if os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2)
	}

	lines := [][]byte{}
	var err error

	if len(os.Args) >= 4 {
		lines, err = readFile(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: read input file: %v\n", err)
			os.Exit(2)
		}
	} else {
		line, err := io.ReadAll(os.Stdin)
		lines = [][]byte{line}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
			os.Exit(2)
		}

	}

	pattern := os.Args[2]

	regexEngine, err := NewRegexEngine(pattern)

	if err != nil {
		log.Fatal("error parsing regex")
	}

	_, match := regexEngine.matchMultiLine(lines)

	if !match {
		os.Exit(1)
	}

}

type RegexEngine struct {
	pattern string
	nfa     NFA
}

func NewRegexEngine(pattern string) (RegexEngine, error) {
	rg := RegexEngine{
		pattern: pattern,
	}

	err := rg.parsePattern()

	if err != nil {
		return RegexEngine{}, err
	}

	return rg, nil
}

func (rg *RegexEngine) parsePattern() error {

	parser := Parser{conversion: Conversion{}, pattern: rg.pattern, pos: 0}
	nfa, err := parser.parse()

	if err != nil {
		return err
	}
	rg.nfa = nfa

	return nil
}

func (rg RegexEngine) matchMultiLine(lines [][]byte) ([][][]byte, bool) {
	match := false
	multiLineMatches := [][][]byte{}
	for _, line := range lines {
		matches := rg.matchLine(line)

		if len(matches) > 0 {
			match = true
			fmt.Printf("%s", line)
		}
		multiLineMatches = append(multiLineMatches, matches)
	}

	return multiLineMatches, match
}

func (rg RegexEngine) matchLine(line []byte) [][]byte {
	isStartAnchor := false
	if len(rg.pattern) > 0 && rg.pattern[0] == '^' {
		isStartAnchor = true
	}

	matches := rg.nfa.findAllMatches(line, isStartAnchor)

	return matches
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
	startGroup  []string
	endGroup    []string
}

var stateCounter int = 0
var capturingGroupCounter int = 1

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

type MemoryGroup struct {
	start int
	end   int
}

type Memory struct {
	activeGroup map[string]MemoryGroup
	groupMatch  map[string]MemoryGroup
}

type Stack struct {
	data   []StackData
	memory Memory
}

func (n *NFA) run(line []byte, index int) (bool, []byte, int) {
	stack := Stack{memory: Memory{activeGroup: make(map[string]MemoryGroup), groupMatch: make(map[string]MemoryGroup)}}
	stack.push(n.states[n.initState], 0)
	line = line[index:]
	for stack.length() > 0 {
		item := stack.pop()
		n.compueGroup(item, &stack, line)
		if item.currentState.isFinal {
			return true, line[:item.i], index + item.i
		}

		for i := len(item.currentState.transitions) - 1; i >= 0; i-- {
			transition := item.currentState.transitions[i]

			if !(item.i < len(line) || transition.matcher.isEpsilon()) {
				continue
			}
			match := transition.matcher.match(line, item.i, stack.memory)

			if match.match {
				newIndex := item.i
				if !transition.matcher.isEpsilon() {
					newIndex += match.consume
				}
				toState := n.states[transition.to]
				stack.push(toState, newIndex)
			}
		}
	}
	return false, nil, 0
}

func (n *NFA) compueGroup(stackdAta StackData, stack *Stack, line []byte) {
	for _, item := range stackdAta.currentState.startGroup {
		stack.memory.activeGroup[item] = MemoryGroup{
			start: stackdAta.i,
		}
	}

	for _, item := range stackdAta.currentState.endGroup {
		stack.memory.groupMatch[item] = MemoryGroup{
			end:   stackdAta.i,
			start: stack.memory.activeGroup[item].start,
		}
	}
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
	union := n.states[unionStateName]
	union.startGroup = append(union.startGroup, nfa.states[nfa.initState].startGroup...)
	union.endGroup = append(union.endGroup, nfa.states[nfa.initState].endGroup...)

	n.states[unionStateName] = union

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
	capturingGroupCounter = 1
	return p.parseAlternation()
}

func (p *Parser) parseAlternation() (NFA, error) {
	// 	          ε                ε
	//       +-------> ( N(s) ) ------->+
	//       |          	            |
	//   -->(q1)                       (q2)<--
	//       |         	 	            |
	//       +-------> ( N(t) ) ------->+
	//           ε                ε

	// 1. Create a new start state q1
	// 2. Add epsilon transition to starting point of left alternation
	// 3. Add epsilion transition to starting point of right alternation
	// 4. Create end state q2
	// 5. Add epsilon transition from ending state of left alternation to end state
	// 6. Add epsilon transition from ending state or right alternation to end state
	// 7. Remove end state from left and right alternation

	left, err := p.parseConcatenation()
	if err != nil {
		return NFA{}, err
	}

	if !p.isEnd() && p.pattern[p.pos] == '|' {
		start := NewState()
		end := NewState()
		end.isFinal = true
		nfa := NFA{states: map[string]State{}}
		nfa.addStates([]State{start, end})
		nfa.setInitState(start)
		nfa.setFinalStates([]State{end})
		for _, state := range left.states {
			if state.isFinal {
				state.isFinal = false
			}
			nfa.addStates([]State{state})
		}
		nfa.addTransition(start, nfa.states[left.initState], EpsilonMatcher{})
		for _, state := range left.endStates {
			nfa.addTransition(nfa.states[state], end, EpsilonMatcher{})
		}

		p.pos++

		right, err := p.parseAlternation()

		if err != nil {
			return NFA{}, err
		}

		for _, state := range right.states {
			if state.isFinal {
				state.isFinal = false
			}
			nfa.addStates([]State{state})
		}
		nfa.addTransition(start, nfa.states[right.initState], EpsilonMatcher{})

		for _, state := range right.endStates {
			nfa.addTransition(nfa.states[state], end, EpsilonMatcher{})
		}

		return nfa, nil

	}

	return left, nil
}

func (p *Parser) parseConcatenation() (NFA, error) {
	left, err := p.parseRepeat()
	if err != nil {
		return NFA{}, err
	}
	for !p.isEnd() && p.pattern[p.pos] != ')' && p.pattern[p.pos] != '|' {
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
		return p.parseCharClass()
	case '^':
		return p.parseCaretAnchor()
	case '$':
		return p.parseDollarAnchor()
	case '.':
		return p.parseDot()
	case '(':
		return p.parseGroup()
	default:
		return p.parseLiteral()
	}
}

func (p *Parser) parseGroup() (NFA, error) {

	//   -->(q1) ------->   ( N(s) )    ------->   (q2)
	// 1. Add new start state (q1)
	// 2. Add end state (q2)
	// 3. Add eppsilon transition from q1 to init state of N(s)
	// 4. Add epsilon transition from ending states of N(s) to q2

	p.pos++
	capturingGroup := strconv.Itoa(capturingGroupCounter)
	capturingGroupCounter++
	nfa, err := p.parseAlternation()
	if p.pattern[p.pos] != ')' {
		return NFA{}, fmt.Errorf("invalid pattern missing closing bracket )")
	}
	p.pos++

	start := NewState()
	start.startGroup = []string{capturingGroup}
	end := NewState()
	end.endGroup = []string{capturingGroup}

	end.isFinal = true

	nfa.addStates([]State{start, end})

	initState := nfa.initState
	endStates := nfa.endStates
	nfa.setInitState(start)
	nfa.setFinalStates([]State{end})

	nfa.addTransition(start, nfa.states[initState], EpsilonMatcher{})
	for _, item := range endStates {
		nfa.addTransition(nfa.states[item], end, EpsilonMatcher{})
		state := nfa.states[item]

		state.isFinal = false
		nfa.states[item] = state
	}

	return nfa, err

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
	}

	if esc >= '1' && esc <= '9' {
		return p.conversion.oneStepNFA(BackreferenceMatcher{groupId: string(esc)})
	}

	return p.parseLiteral()
}

func (p *Parser) parseCharClass() (NFA, error) {
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

type MatchResult struct {
	match   bool
	consume int
}
type Matcher interface {
	match(b []byte, index int, memory Memory) MatchResult
	isEpsilon() bool
}

type EpsilonMatcher struct {
	char byte
}

func (lm EpsilonMatcher) match(b []byte, index int, memory Memory) MatchResult {
	return MatchResult{match: true, consume: 0}
}

func (lm EpsilonMatcher) isEpsilon() bool {
	return true
}

type LiteralMatcher struct {
	char byte
}

func (lm LiteralMatcher) match(b []byte, index int, memory Memory) MatchResult {
	return MatchResult{match: lm.char == b[index], consume: 1}
}

func (lm LiteralMatcher) isEpsilon() bool {
	return false
}

type DigitMatcher struct {
	ranges []CharRange
}

func (dm DigitMatcher) match(b []byte, index int, memory Memory) MatchResult {

	return MatchResult{match: matchRanges(dm.ranges, b[index]), consume: 1}
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

func (wm WordMatcher) match(b []byte, index int, memory Memory) MatchResult {
	return MatchResult{match: matchRanges(wm.ranges, b[index]) || matchChars(wm.chars, b[index]), consume: 1}
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

func (cgm CharacterGroupMatcher) match(b []byte, index int, memory Memory) MatchResult {
	base := matchChars(cgm.chars, b[index]) || matchRanges(cgm.ranges, b[index])

	if cgm.isNegative {
		return MatchResult{match: !base, consume: 1}
	}

	return MatchResult{match: base, consume: 1}
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

func (startOfStringMatcher StartOfStringMatcher) match(b []byte, index int, memory Memory) MatchResult {
	return MatchResult{match: index == 0, consume: 0}
}

func (lm StartOfStringMatcher) isEpsilon() bool {
	return true
}

type EndOfStringMatcher struct{}

func (endOfStringMatcher EndOfStringMatcher) match(b []byte, index int, memory Memory) MatchResult {
	return MatchResult{match: len(b) == index, consume: 0}
}

func (lm EndOfStringMatcher) isEpsilon() bool {
	return true
}

type AnyCharMatcher struct{}

func (anyCharMatcher AnyCharMatcher) match(b []byte, index int, memory Memory) MatchResult {
	return MatchResult{match: true, consume: 1}
}

func (anyCharMatcher AnyCharMatcher) isEpsilon() bool {
	return false
}

type BackreferenceMatcher struct {
	groupId string
}

func (backreferenceMatcher BackreferenceMatcher) match(line []byte, index int, memory Memory) MatchResult {
	memGroup, ok := memory.groupMatch[backreferenceMatcher.groupId]

	if !ok {
		panic("memory doesnt have group" + backreferenceMatcher.groupId)
	}
	i := index

	for _, b := range line[memGroup.start:memGroup.end] {
		if byte(b) != line[i] {
			return MatchResult{match: false, consume: i - index}
		}
		i++
	}

	return MatchResult{match: true, consume: i - index}

}

func (backreferenceMatcher BackreferenceMatcher) isEpsilon() bool {
	return false
}
