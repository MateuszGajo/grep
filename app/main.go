package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func bytesToStrings(bs [][]byte) []string {
	var ss []string
	for _, b := range bs {
		ss = append(ss, string(b))
	}
	return ss
}

func main() {
	web := flag.Bool("web", false, "start web server")
	flag.Parse()

	if *web {
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

	matches, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if len(matches) == 0 {
		os.Exit(1)
	}

}
func matchLineDetails(line []byte, pattern string) ([][]byte, NFA, error) {
	parser := Parser{conversion: Conversion{}, pattern: pattern, pos: 0}
	nfa, err := parser.parse()

	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	matches := nfa.findAllMatches(line)
	fmt.Println("matches??", matches)
	fmt.Println("nfa")
	fmt.Printf("%+v \n", nfa)

	return matches, nfa, err
}

func matchLine(line []byte, pattern string) ([][]byte, error) {
	parser := Parser{conversion: Conversion{}, pattern: pattern, pos: 0}
	nfa, err := parser.parse()

	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	matches := nfa.findAllMatches(line)
	fmt.Println("matches??", matches)

	return matches, err
}

type NFATransition struct {
	to        string
	matcher   Matcher
	isEpsilon bool
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

type Stack struct {
	data []StackData
}

func (n *NFA) run(line []byte) (bool, []byte) {
	stack := Stack{}
	stack.push(n.states[n.initState], 0)

	for stack.length() > 0 {
		item := stack.pop()
		if item.currentState.isFinal {
			return true, line[:item.i]
		}

		for _, transition := range item.currentState.transitions {
			fmt.Println("line", line)
			fmt.Println("item", item.i)
			if item.i < len(line) && transition.matcher.match(line[item.i]) {
				newIndex := item.i
				if !transition.isEpsilon {
					newIndex += 1
				}
				toState := n.states[transition.to]
				stack.push(toState, newIndex)
			}
		}
	}
	return false, nil
}

func (n *NFA) findAllMatches(input []byte) [][]byte {
	matches := [][]byte{}
	for i := 0; i < len(input); i++ {
		ok, match := n.run(input[i:])
		if ok {
			matches = append(matches, match)
		}

	}

	return matches
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

func (n *NFA) appendNfa(nfa NFA, unionStateName string) {
	// first nfa is q1 ->(\d) q2
	//second nfa is q3 ->(\d) q4
	// as result we need to have q1->q2->q4, q3= q2 state basically
	// need also set q4 as ending state, and remove q2 as ending state
	// pass transition from q3 to q2
	states := []State{}
	for _, item := range nfa.states {
		// skip q3 state
		fmt.Println(item)
		fmt.Println(unionStateName)
		if item.name == nfa.initState {
			continue
		}
		states = append(states, item)
	}

	unionState := n.states[unionStateName]
	unionState.isFinal = false
	n.states[unionStateName] = unionState
	n.addStates(states)

	nfaInitState := nfa.states[nfa.initState]
	for _, transition := range nfaInitState.transitions {
		n.addTransition(unionState, n.states[transition.to], transition.matcher)
	}

	newEndStates := []string{}

	for _, item := range n.endStates {

		if item == unionStateName {
			newEndStates = append(newEndStates, nfa.endStates...)
		} else {
			newEndStates = append(newEndStates, item)
		}
	}
	n.endStates = newEndStates

}

type Conversion struct {
}

func (c Conversion) oneStepNFA(matcher Matcher) (NFA, error) {
	nfa := NFA{states: map[string]State{}}
	a := *NewState()
	b := *NewState()
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

	return leftAtom, err
	// rightAtom, err := p.parseAtom()

	// leftAtom.appendNfa(rightAtom, leftAtom.endStates[0])

	// return leftAtom, err
}

func (p *Parser) parseAtom() (NFA, error) {
	switch p.pattern[p.pos] {
	case '\\':
		return p.parseEscape(p.pattern)
	default:
		return p.parseLiteral(p.pattern)
	}
}

func (p *Parser) parseEscape(pattern string) (NFA, error) {
	p.pos++
	esc := p.pattern[p.pos]
	p.pos++
	switch esc {
	case 'd':
		return p.conversion.oneStepNFA(DigitMatcher{})
	default:
		return NFA{}, fmt.Errorf("unsupported escape: \\%c", esc)
	}
}

func (p *Parser) parseLiteral(pattern string) (NFA, error) {
	matcher := LiteralMatcher{char: pattern[p.pos]}
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
	match(b byte) bool
}

type LiteralMatcher struct {
	char byte
}

func (lm LiteralMatcher) match(b byte) bool {
	return lm.char == b
}

type DigitMatcher struct {
}

func (dm DigitMatcher) match(b byte) bool {
	if b >= '0' && b <= '9' {
		return true
	}

	return false
}
