package symbolic_regression

import (
	"fmt"
	"strconv"
	"strings"
)

// FormulaFunction is a function that evaluates a formula given training inputs
type FormulaFunction func(inputs TrainingInputs) float64

// ParseFormula parses a formula string into a Node tree
// Supports: +, -, *, /, sqrt(), log(), exp(), abs(), pow(), max(), min()
// Variables: cagr, score, regime, long_term, fundamentals, etc.
func ParseFormula(formulaStr string) (*Node, error) {
	formulaStr = strings.TrimSpace(formulaStr)
	if formulaStr == "" {
		return nil, fmt.Errorf("empty formula string")
	}

	// Parse using recursive descent
	tokens := tokenize(formulaStr)
	parser := &formulaParser{tokens: tokens, pos: 0}

	node, err := parser.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("failed to parse formula: %w", err)
	}

	if parser.pos < len(parser.tokens) {
		return nil, fmt.Errorf("unexpected token at end of formula")
	}

	return node, nil
}

// tokenize tokenizes a formula string
func tokenize(formula string) []token {
	var tokens []token
	i := 0

	for i < len(formula) {
		// Skip whitespace
		if isWhitespace(formula[i]) {
			i++
			continue
		}

		// Check for operators
		if i+1 < len(formula) {
			if formula[i:i+2] == "**" {
				tokens = append(tokens, token{typ: tokenPower, value: "**"})
				i += 2
				continue
			}
		}

		switch formula[i] {
		case '+':
			tokens = append(tokens, token{typ: tokenPlus, value: "+"})
			i++
		case '-':
			tokens = append(tokens, token{typ: tokenMinus, value: "-"})
			i++
		case '*':
			tokens = append(tokens, token{typ: tokenMultiply, value: "*"})
			i++
		case '/':
			tokens = append(tokens, token{typ: tokenDivide, value: "/"})
			i++
		case '(':
			tokens = append(tokens, token{typ: tokenLParen, value: "("})
			i++
		case ')':
			tokens = append(tokens, token{typ: tokenRParen, value: ")"})
			i++
		case ',':
			tokens = append(tokens, token{typ: tokenComma, value: ","})
			i++
		default:
			// Try to parse number or identifier
			start := i
			if isDigit(formula[i]) || formula[i] == '.' {
				// Number
				for i < len(formula) && (isDigit(formula[i]) || formula[i] == '.' || formula[i] == 'e' || formula[i] == 'E' || formula[i] == '+' || formula[i] == '-') {
					i++
				}
				tokens = append(tokens, token{typ: tokenNumber, value: formula[start:i]})
			} else if isLetter(formula[i]) {
				// Identifier (variable or function)
				for i < len(formula) && (isLetter(formula[i]) || isDigit(formula[i]) || formula[i] == '_') {
					i++
				}
				tokens = append(tokens, token{typ: tokenIdentifier, value: formula[start:i]})
			} else {
				// Unknown character - skip
				i++
			}
		}
	}

	return tokens
}

// token types
type tokenType int

const (
	tokenNumber tokenType = iota
	tokenIdentifier
	tokenPlus
	tokenMinus
	tokenMultiply
	tokenDivide
	tokenPower
	tokenLParen
	tokenRParen
	tokenComma
	tokenEOF
)

type token struct {
	typ   tokenType
	value string
}

// formulaParser is a recursive descent parser
type formulaParser struct {
	tokens []token
	pos    int
}

func (p *formulaParser) parseExpression() (*Node, error) {
	return p.parseAdditive()
}

func (p *formulaParser) parseAdditive() (*Node, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokenPlus {
			p.pos++
			right, err := p.parseMultiplicative()
			if err != nil {
				return nil, err
			}
			left = &Node{Type: NodeTypeOperation, Op: OpAdd, Left: left, Right: right}
		} else if tok.typ == tokenMinus {
			p.pos++
			right, err := p.parseMultiplicative()
			if err != nil {
				return nil, err
			}
			left = &Node{Type: NodeTypeOperation, Op: OpSubtract, Left: left, Right: right}
		} else {
			break
		}
	}

	return left, nil
}

func (p *formulaParser) parseMultiplicative() (*Node, error) {
	left, err := p.parsePower()
	if err != nil {
		return nil, err
	}

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokenMultiply {
			p.pos++
			right, err := p.parsePower()
			if err != nil {
				return nil, err
			}
			left = &Node{Type: NodeTypeOperation, Op: OpMultiply, Left: left, Right: right}
		} else if tok.typ == tokenDivide {
			p.pos++
			right, err := p.parsePower()
			if err != nil {
				return nil, err
			}
			left = &Node{Type: NodeTypeOperation, Op: OpDivide, Left: left, Right: right}
		} else {
			break
		}
	}

	return left, nil
}

func (p *formulaParser) parsePower() (*Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	if p.pos < len(p.tokens) && p.tokens[p.pos].typ == tokenPower {
		p.pos++
		right, err := p.parsePower()
		if err != nil {
			return nil, err
		}
		return &Node{Type: NodeTypeOperation, Op: OpPower, Left: left, Right: right}, nil
	}

	return left, nil
}

func (p *formulaParser) parseUnary() (*Node, error) {
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of expression")
	}

	tok := p.tokens[p.pos]
	if tok.typ == tokenMinus {
		p.pos++
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &Node{Type: NodeTypeOperation, Op: OpNegate, Left: operand}, nil
	}

	return p.parsePrimary()
}

func (p *formulaParser) parsePrimary() (*Node, error) {
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of expression")
	}

	tok := p.tokens[p.pos]

	if tok.typ == tokenNumber {
		p.pos++
		val, err := strconv.ParseFloat(tok.value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", tok.value)
		}
		return &Node{Type: NodeTypeConstant, Value: val}, nil
	}

	if tok.typ == tokenIdentifier {
		p.pos++
		name := tok.value

		// Check if it's a function call
		if p.pos < len(p.tokens) && p.tokens[p.pos].typ == tokenLParen {
			return p.parseFunction(name)
		}

		// It's a variable
		return &Node{Type: NodeTypeVariable, Variable: name}, nil
	}

	if tok.typ == tokenLParen {
		p.pos++
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.pos >= len(p.tokens) || p.tokens[p.pos].typ != tokenRParen {
			return nil, fmt.Errorf("expected ')'")
		}
		p.pos++
		return expr, nil
	}

	return nil, fmt.Errorf("unexpected token: %s", tok.value)
}

func (p *formulaParser) parseFunction(name string) (*Node, error) {
	// Skip '('
	p.pos++

	// Parse arguments
	var args []*Node
	for {
		if p.pos >= len(p.tokens) {
			return nil, fmt.Errorf("unexpected end in function call")
		}

		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		if p.pos >= len(p.tokens) {
			return nil, fmt.Errorf("unexpected end in function call")
		}

		if p.tokens[p.pos].typ == tokenRParen {
			p.pos++
			break
		}

		if p.tokens[p.pos].typ != tokenComma {
			return nil, fmt.Errorf("expected ',' or ')' in function call")
		}
		p.pos++
	}

	// Map function name to operation
	switch name {
	case "sqrt":
		if len(args) != 1 {
			return nil, fmt.Errorf("sqrt requires 1 argument")
		}
		return &Node{Type: NodeTypeOperation, Op: OpSqrt, Left: args[0]}, nil
	case "log":
		if len(args) != 1 {
			return nil, fmt.Errorf("log requires 1 argument")
		}
		return &Node{Type: NodeTypeOperation, Op: OpLog, Left: args[0]}, nil
	case "exp":
		if len(args) != 1 {
			return nil, fmt.Errorf("exp requires 1 argument")
		}
		return &Node{Type: NodeTypeOperation, Op: OpExp, Left: args[0]}, nil
	case "abs":
		if len(args) != 1 {
			return nil, fmt.Errorf("abs requires 1 argument")
		}
		return &Node{Type: NodeTypeOperation, Op: OpAbs, Left: args[0]}, nil
	case "pow":
		if len(args) != 2 {
			return nil, fmt.Errorf("pow requires 2 arguments")
		}
		return &Node{Type: NodeTypeOperation, Op: OpPower, Left: args[0], Right: args[1]}, nil
	case "max":
		if len(args) != 2 {
			return nil, fmt.Errorf("max requires 2 arguments")
		}
		return &Node{Type: NodeTypeOperation, Op: OpMax, Left: args[0], Right: args[1]}, nil
	case "min":
		if len(args) != 2 {
			return nil, fmt.Errorf("min requires 2 arguments")
		}
		return &Node{Type: NodeTypeOperation, Op: OpMin, Left: args[0], Right: args[1]}, nil
	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}

// FormulaToFunction converts a Node tree to an executable function
func FormulaToFunction(formula *Node) FormulaFunction {
	return func(inputs TrainingInputs) float64 {
		variables := getVariableMap(inputs)
		return formula.Evaluate(variables)
	}
}

// Helper functions
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
