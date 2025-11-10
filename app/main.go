package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

type File struct {
	name string
	data [][]byte
}

func readFiles(filenames []string) ([]File, error) {
	files := []File{}
	for _, filename := range filenames {
		data, err := readFile(filename)

		if err != nil {
			return nil, err
		}
		files = append(files, File{name: filename, data: data})
	}

	return files, nil
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

func listfilePath(folder string) ([]string, error) {
	filePath := []string{}
	err := filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			filePath = append(filePath, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return filePath, nil
}

type Args struct {
	pattern     string
	isRecusrive bool
	directory   string
	filePathes  []string
}

func parseArgs() (Args, error) {
	args := Args{}
	argsCopy := os.Args[1:]

	if argsCopy[0] == "-r" {
		args.isRecusrive = true
		argsCopy = argsCopy[1:]
	}

	if argsCopy[0] != "-E" {
		return Args{}, fmt.Errorf("usage: mygrep -E <pattern>")
	}
	argsCopy = argsCopy[1:]

	args.pattern = argsCopy[0]
	argsCopy = argsCopy[1:]

	if args.isRecusrive {
		args.directory = argsCopy[0]
		return args, nil
	}

	for i := 0; i < len(argsCopy); i++ {
		args.filePathes = append(args.filePathes, argsCopy[i])
	}

	return args, nil
}

func cli() {
	args, err := parseArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	if args.isRecusrive {
		filepath, err := listfilePath(args.directory)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		args.filePathes = filepath

	}

	regexEngine, err := NewRegexEngine(args.pattern)

	if err != nil {
		log.Fatal("error parsing regex")
	}

	if len(args.filePathes) > 0 {
		files, err := readFiles(args.filePathes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: read input file: %v\n", err)
			os.Exit(2)
		}

		isAnyMatch := false
		isPrefix := len(args.filePathes) > 1
		for _, file := range files {
			matches, isMatch := regexEngine.matchMultiLine(file.data)
			if isMatch {
				isAnyMatch = true
			}

			for _, match := range matches {
				if isPrefix {
					fmt.Println(file.name + ":" + string(match.line))

				} else {
					fmt.Println(string(match.line))
				}
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: read input file: %v\n", err)
			os.Exit(2)
		}
		if !isAnyMatch {
			os.Exit(1)
		}
	} else {
		line, err := io.ReadAll(os.Stdin)
		lines := [][]byte{line}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
			os.Exit(2)
		}
		_, isMatch := regexEngine.matchMultiLine(lines)

		if !isMatch {
			os.Exit(1)
		}

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

// func (rg RegexEngine) matchFilesMultiline(files []File) ([][][]byte, bool) {

// }

type RegexOutput struct {
	line        []byte
	matchPhrase [][]byte
}

func (rg RegexEngine) matchMultiLine(lines [][]byte) ([]RegexOutput, bool) {
	match := false
	multiLineMatches := []RegexOutput{}
	for _, line := range lines {
		matches := rg.matchLine(line)

		if len(matches) > 0 {
			match = true
			multiLineMatches = append(multiLineMatches, RegexOutput{line: line, matchPhrase: matches})
		}

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
	to      State
	matcher Matcher
}

type NFA struct {
	initState State
	endStates []*State
	states    []State
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
	n.states = append(n.states, states...)
}

func (n *NFA) findState(name string) *State {
	for i, _ := range n.states {
		if n.states[i].name == name {
			return &n.states[i]
		}
	}
	return nil
}

func (n *NFA) addTransition(from State, to *State, matcher Matcher) error {

	fromState := n.findState(from.name)

	fromState.transitions = append(fromState.transitions, NFATransition{to: *to, matcher: matcher})

	return nil
}

func (n *NFA) setInitState(state *State) {
	n.initState = *state
}

func (n *NFA) setFinalStates(states []*State) error {
	n.endStates = []*State{}
	for _, state := range states {
		state := n.findState(state.name)

		state.isFinal = true

		n.endStates = append(n.endStates, state)
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
	stack.push(*n.findState(n.initState.name), 0)
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
				toState := n.findState(transition.to.name)
				stack.push(*toState, newIndex)
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
		if item.name == nfa.initState.name {
			continue
		}
		n.states = append(n.states, item)
	}

	for _, transition := range nfa.findState(nfa.initState.name).transitions {
		n.addTransition(*n.findState(unionStateName), n.findState(transition.to.name), transition.matcher)
	}
	union := n.findState(unionStateName)
	union.startGroup = append(union.startGroup, nfa.findState(nfa.initState.name).startGroup...)
	union.endGroup = append(union.endGroup, nfa.findState(nfa.initState.name).endGroup...)

	newEndStates := []*State{}

	for _, item := range n.endStates {

		if item.name == unionStateName {
			newEndStates = append(newEndStates, nfa.endStates...)
			unionState := n.findState(unionStateName)
			unionState.isFinal = false
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
	nfa := NFA{states: []State{}}
	a := NewState()
	b := NewState()
	nfa.addStates([]State{a, b})
	nfa.setInitState(&a)
	err := nfa.setFinalStates([]*State{&b})
	if err != nil {
		return nfa, err
	}
	err = nfa.addTransition(a, &b, matcher)

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
		nfa := NFA{states: []State{}}
		nfa.addStates([]State{start, end})
		nfa.setInitState(&start)
		nfa.setFinalStates([]*State{&end})
		for _, state := range left.states {
			if state.isFinal {
				state.isFinal = false
			}
			nfa.addStates([]State{state})
		}
		nfa.addTransition(start, nfa.findState(left.initState.name), EpsilonMatcher{})
		for _, state := range left.endStates {
			nfa.addTransition(*nfa.findState(state.name), &end, EpsilonMatcher{})
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
		nfa.addTransition(start, nfa.findState(right.initState.name), EpsilonMatcher{})

		for _, state := range right.endStates {
			nfa.addTransition(*nfa.findState(state.name), &end, EpsilonMatcher{})
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
		left.appendNfa(right, left.endStates[0].name)

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
		leftAtom.setInitState(&q1)
		leftAtom.addTransition(q1, leftAtom.findState(currentInitState.name), EpsilonMatcher{})
		// Greedy matcher, need to be added first
		leftAtom.addTransition(*leftAtom.findState(leftAtom.endStates[0].name), leftAtom.findState(currentInitState.name), EpsilonMatcher{})
		leftAtom.addTransition(*leftAtom.findState(leftAtom.endStates[0].name), &q4, EpsilonMatcher{})

		last := leftAtom.findState(leftAtom.endStates[0].name)
		last.isFinal = false

		leftAtom.setFinalStates([]*State{&q4})
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
		leftAtom.setInitState(&q1)
		leftAtom.addTransition(q1, leftAtom.findState(currentInitState.name), EpsilonMatcher{})
		leftAtom.addTransition(q1, &q4, EpsilonMatcher{})
		// Greedy matcher, need to be added first
		leftAtom.addTransition(*leftAtom.findState(leftAtom.endStates[0].name), &q4, EpsilonMatcher{})

		last := leftAtom.findState(leftAtom.endStates[0].name)
		last.isFinal = false

		leftAtom.setFinalStates([]*State{&q4})
		p.pos++
	case '*':
		/*
					  ┌────────ε────────┐
					  ▼          	    │
			(q1) -ε> (q2) -condition-> (q3) -ε> ((q4))
			  │                    			       ▲
			  └────────────────ε───────────────────┘

			1. Create start state q1
			2. Create end state q4
			3. Add epislon transition from q1 to q4
			4. Add epsilon transition from q1 to q2
			5. Add epsilon transition from q3 to q2 (loop repetition), by default it's greedy so prioritize it rather than exit the loop
			6. Add epsilon transition from q3 to q4
		*/
		q1 := NewState()
		q4 := NewState()
		q4.isFinal = true

		leftAtom.addStates([]State{q1, q4})
		currentInitState := leftAtom.initState
		leftAtom.setInitState(&q1)
		leftAtom.addTransition(q1, leftAtom.findState(currentInitState.name), EpsilonMatcher{})
		leftAtom.addTransition(q1, &q4, EpsilonMatcher{})
		leftAtom.addTransition(*leftAtom.findState(leftAtom.endStates[0].name), leftAtom.findState(currentInitState.name), EpsilonMatcher{})
		leftAtom.addTransition(*leftAtom.findState(leftAtom.endStates[0].name), &q4, EpsilonMatcher{})

		last := leftAtom.findState(leftAtom.endStates[0].name)
		last.isFinal = false

		leftAtom.setFinalStates([]*State{&q4})
		p.pos++
		// case '{':
		// 	p.pos++
		// 	currByte := p.pattern[p.pos]
		// 	numberString := ""
		// 	for currByte >= '0' && currByte <= '9' {
		// 		numberString += string(currByte)
		// 		p.pos++
		// 	}

		// 	num, err := strconv.Atoi(numberString)

		// 	if err != nil {
		// 		panic("should never happend we read something wrong")
		// 	}

		// 	if currByte != '}' {
		// 		panic("expected end of quantifier, currently support only for exact match e.g. {2}")
		// 	}
		// 	/*
		// 		(q1) -condition-> (q2)
		// 	*/
		// 	// we need to repeat this multiple times
		// 	// leftAtom.states[leftAtom.endStates[0]].transitions

		// 	for i := 0; i < num; i++ {

		// 		// the simplest way possible is leftatom.end should have transition to leftatom.start n times
		// 	}

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
	nfa.setInitState(&start)
	nfa.setFinalStates([]*State{&end})
	nfa.addTransition(start, nfa.findState(initState.name), EpsilonMatcher{})
	for _, item := range endStates {
		state := nfa.findState(item.name)
		nfa.addTransition(*state, &end, EpsilonMatcher{})

		state.isFinal = false
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
