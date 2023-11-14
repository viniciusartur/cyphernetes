package cmd

import (
	"log"
	"strings"
	"text/scanner"
)

type Token int

const (
	ILLEGAL Token = iota
	WS
)

type Lexer struct {
	s   scanner.Scanner
	buf struct {
		tok Token
		lit string
	}
	definingProps bool
}

func NewLexer(input string) *Lexer {
	var s scanner.Scanner
	s.Init(strings.NewReader(input))
	s.Whitespace = 1<<'\t' | 1<<'\r' | 1<<' '
	return &Lexer{s: s}
}

func (l *Lexer) Lex(lval *yySymType) int {
	debugLog("Lexing... ", l.s.Peek(), " (", string(l.s.Peek()), ")")
	if l.buf.tok == EOF { // If we have already returned EOF, keep returning EOF
		logDebug("Zero (buffered EOF)")
		return 0
	}

	// Check if we are capturing a JSONPATH
	if l.buf.tok == RETURN || l.buf.tok == LBRACE || (l.buf.tok == COMMA && l.definingProps) {
		lval.strVal = ""
		// Consume and ignore any whitespace
		ch := l.s.Peek()
		for ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.s.Next() // Consume the whitespace
			ch = l.s.Peek()
		}

		// Capture the JSONPATH
		for isValidJsonPathChar(ch) {
			l.s.Next() // Consume the character
			lval.strVal += string(ch)
			ch = l.s.Peek()
		}

		l.buf.tok = ILLEGAL // Indicate that we've read a JSONPATH.
		logDebug("Returning JSONPATH token with value:", lval.strVal)
		return int(JSONPATH)
	}

	// Handle normal tokens
	tok := l.s.Scan()
	logDebug("Scanned token:", tok)
	logDebug("Token text:", string(tok))

	switch tok {
	case scanner.Ident:
		lit := l.s.TokenText()
		if strings.ToUpper(lit) == "MATCH" {
			logDebug("Returning MATCH token")
			return int(MATCH)
		} else if strings.ToUpper(lit) == "RETURN" {
			l.buf.tok = RETURN // Indicate that we've read a RETURN.
			logDebug("Returning RETURN token")
			return int(RETURN)
		} else if strings.ToUpper(lit) == "TRUE" || strings.ToUpper(lit) == "FALSE" {
			lval.strVal = l.s.TokenText()
			logDebug("Returning BOOLEAN token with value:", lval.strVal)
			return int(BOOLEAN)
		} else {
			lval.strVal = lit
			logDebug("Returning IDENT token with value:", lval.strVal)
			return int(IDENT)
		}
	case scanner.EOF:
		logDebug("Returning EOF token")
		l.buf.tok = EOF // Indicate that we've read an EOF.
		return int(EOF)
	case '(':
		logDebug("Returning LPAREN token")
		l.buf.tok = LPAREN // Indicate that we've read a LPAREN.
		return int(LPAREN)
	case ':':
		l.definingProps = true // Indicate that we've read a COLON.
		logDebug("Returning COLON token")
		return int(COLON)
	case ')':
		logDebug("Returning RPAREN token")
		l.definingProps = false // Indicate that we've read a RPAREN.
		return int(RPAREN)
	case ' ', '\t', '\r':
		logDebug("Ignoring whitespace")
		return int(WS) // Ignore whitespace.
	case '{':
		// Capture a JSON object
		l.buf.tok = LBRACE // Indicate that we've read a LBRACE.
		logDebug("Returning LBRACE token")
		return int(LBRACE)
	case '}':
		logDebug("Returning RBRACE token")
		return int(RBRACE)
	case -6: // QUOTE
		lval.strVal = l.s.TokenText()
		logDebug("Returning STRING token with value:", lval.strVal)
		return int(STRING)
	case scanner.Int:
		lval.strVal = l.s.TokenText()
		logDebug("Returning INT token with value:", lval.strVal)
		return int(INT)
	case ',':
		logDebug("Returning COMMA token")
		l.buf.tok = COMMA // Indicate that we've read a COMMA.
		return int(COMMA)
	default:
		logDebug("Illegal token:", tok)
		return int(ILLEGAL)
	}
}

// Helper function to check if a character is valid in a jsonPath
func isValidJsonPathChar(tok rune) bool {
	// convert to string for easier comparison
	char := string(tok)

	return char == "." || char == "[" || char == "]" ||
		(char >= "0" && char <= "9") || char == "_" ||
		(char >= "a" && char <= "z") || (char >= "A" && char <= "Z") ||
		char == "\"" || char == "*" || char == "$" || char == "#"
}

func (l *Lexer) Error(e string) {
	log.Printf("Error: %v\n", e)
}

type ASTNode struct {
	Name string
	Kind string
}

func NewASTNode(name, kind string) *ASTNode {
	return &ASTNode{Name: name, Kind: kind}
}

func logDebug(v ...interface{}) {
	if logLevel == "debug" {
		log.Println(v...)
	}
}
