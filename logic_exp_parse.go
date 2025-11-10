package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type TokenType int

type Token struct {
	Type  TokenType
	Value string
}

const (
	TOKEN_EOF TokenType = iota
	TOKEN_IDENT
	TOKEN_BOOL
	TOKEN_OPERATOR
	TOKEN_LPAREN
	TOKEN_RPAREN
)

func isValidVarName(name string) bool {
	if name == "" {
		return false
	}
	for _, char := range name {
		if char < 'a' || char > 'z' {
			return false
		}
	}

	keywords := map[string]bool{
		"and":   true,
		"or":    true,
		"not":   true,
		"xor":   true,
		"true":  true,
		"false": true,
	}

	return !keywords[name]
}

func isLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z'
}

func getWordType(word string) TokenType {
	operators := map[string]bool{
		"and": true, "or": true, "not": true, "xor": true,
	}

	booleans := map[string]bool{
		"True": true, "False": true,
	}

	if operators[word] {
		return TOKEN_OPERATOR
	}

	if booleans[word] {
		return TOKEN_BOOL
	}

	if isValidVarName(word) {
		return TOKEN_IDENT
	}

	return TOKEN_EOF
}

func parseDeclaration(line string) (string, bool, error) {
	line = strings.ReplaceAll(line, " ", "")

	if !strings.Contains(line, "=") || !strings.HasSuffix(line, ";") {
		return "", false, fmt.Errorf("error")
	}

	parts := strings.Split(line, "=")
	if len(parts) != 2 {
		return "", false, fmt.Errorf("error")
	}

	varName := parts[0]
	valueStr := strings.TrimSuffix(parts[1], ";")

	if !isValidVarName(varName) {
		return "", false, fmt.Errorf("error")
	}

	var value bool
	switch valueStr {
	case "True":
		value = true
	case "False":
		value = false
	default:
		return "", false, fmt.Errorf("error")
	}

	return varName, value, nil
}

func lexer(expression string) ([]Token, error) {
	var tokens []Token
	i := 0
	n := len(expression)

	for i < n {
		if expression[i] == ' ' {
			i++
			continue
		}

		if expression[i] == '(' {
			tokens = append(tokens, Token{TOKEN_LPAREN, "("})
			i++
			continue
		}

		if expression[i] == ')' {
			tokens = append(tokens, Token{TOKEN_RPAREN, ")"})
			i++
			continue
		}

		if isLetter(expression[i]) {
			start := i
			for i < n && isLetter(expression[i]) {
				i++
			}
			word := expression[start:i]
			tokenType := getWordType(word)
			if tokenType == TOKEN_EOF {
				return nil, fmt.Errorf("error")
			}

			tokens = append(tokens, Token{tokenType, word})
			continue
		}
		return nil, fmt.Errorf("error")
	}
	return tokens, nil
}

func parseExpression(tokens []Token, vars map[string]bool, pos *int) (bool, error) {
	return parseOr(tokens, vars, pos)
}

func parseOr(tokens []Token, vars map[string]bool, pos *int) (bool, error) {
	left, err := parseXor(tokens, vars, pos)
	if err != nil {
		return false, err
	}

	for *pos < len(tokens) && tokens[*pos].Value == "or" {
		*pos++
		right, err := parseXor(tokens, vars, pos)
		if err != nil {
			return false, err
		}
		left = left || right
	}
	return left, nil
}

func parseXor(tokens []Token, vars map[string]bool, pos *int) (bool, error) {
	left, err := parseAnd(tokens, vars, pos)
	if err != nil {
		return false, err
	}

	for *pos < len(tokens) && tokens[*pos].Value == "xor" {
		*pos++
		right, err := parseAnd(tokens, vars, pos)
		if err != nil {
			return false, err
		}
		left = (left && !right) || (!left && right)
	}
	return left, nil
}

func parseAnd(tokens []Token, vars map[string]bool, pos *int) (bool, error) {
	left, err := parseNot(tokens, vars, pos)
	if err != nil {
		return false, err
	}

	for *pos < len(tokens) && tokens[*pos].Value == "and" {
		*pos++
		right, err := parseNot(tokens, vars, pos)
		if err != nil {
			return false, err
		}
		left = left && right
	}
	return left, nil
}

func parseNot(tokens []Token, vars map[string]bool, pos *int) (bool, error) {
	if *pos < len(tokens) && tokens[*pos].Value == "not" {
		*pos++
		operand, err := parseNot(tokens, vars, pos)
		return !operand, err
	}
	return parsePrimary(tokens, vars, pos)
}

func parsePrimary(tokens []Token, vars map[string]bool, pos *int) (bool, error) {
	if *pos >= len(tokens) {
		return false, fmt.Errorf("error")
	}

	token := tokens[*pos]

	switch token.Type {
	case TOKEN_IDENT:
		*pos++
		value, exist := vars[token.Value]
		if !exist {
			return false, fmt.Errorf("error")
		}
		return value, nil

	case TOKEN_BOOL:
		*pos++
		return token.Value == "True", nil

	case TOKEN_LPAREN:
		*pos++
		result, err := parseExpression(tokens, vars, pos)
		if err != nil {
			return false, err
		}

		if *pos >= len(tokens) || tokens[*pos].Type != TOKEN_RPAREN {
			return false, fmt.Errorf("error")
		}
		*pos++
		return result, nil

	default:
		return false, fmt.Errorf("error")
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	variables := make(map[string]bool)
	var expression string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.Contains(line, "=") && strings.HasSuffix(line, ";") {
			name, value, err := parseDeclaration(line)
			if err != nil {
				fmt.Println("[error]")
				return
			}
			variables[name] = value
		} else {
			expression = line
			break
		}
	}

	if expression == "" {
		fmt.Println("[error]")
		return
	}

	tokens, err := lexer(expression)
	if err != nil {
		fmt.Println("[error]")
		return
	}

	for _, token := range tokens {
		if token.Type == TOKEN_IDENT {
			if _, exists := variables[token.Value]; !exists {
				fmt.Println("[error]")
				return
			}
		}
	}

	pos := 0
	result, err := parseExpression(tokens, variables, &pos)
	if err != nil {
		fmt.Println("[error]")
		return
	}

	if pos != len(tokens) {
		fmt.Println("[error]")
		return
	}

	if result {
		fmt.Println("True")
	} else {
		fmt.Println("False")
	}
}
