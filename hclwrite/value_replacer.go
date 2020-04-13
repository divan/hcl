package hclwrite

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Replacer implements token-level value replacer.
type Replacer struct {
	Name  string
	Value Tokens

	depth int // current token depth in nested structures, used to ignore nested values

	inside  bool
	waitFor hclsyntax.TokenType

	// we store matching token pairs counters here
	// ([], (), {}, "", here docs, etc)
	// to skip all internal tokens when looking for the closing one
	matching map[hclsyntax.TokenType]int
}

// NewReplacer returns and inits new Replacer.
func NewReplacer(name string, value Tokens) *Replacer {
	return &Replacer{
		Name:  name,
		Value: value,

		matching: map[hclsyntax.TokenType]int{},
	}
}

func (r *Replacer) processToken(token Token, depth string) (Tokens, bool) {
	// it's the first token after '='. let's figure out what to wait for
	processMatching := func(cTok hclsyntax.TokenType) {
		if r.waitFor == hclsyntax.TokenEOF {
			r.waitFor = cTok
		}
		r.matching[cTok]++
		r.depth++
	}

	switch token.Type {
	case hclsyntax.TokenOQuote:
		processMatching(hclsyntax.TokenCQuote)
	case hclsyntax.TokenOBrack:
		processMatching(hclsyntax.TokenCBrack)
	case hclsyntax.TokenOBrace:
		processMatching(hclsyntax.TokenOBrace)
	case hclsyntax.TokenOHeredoc:
		processMatching(hclsyntax.TokenOHeredoc)
	case hclsyntax.TokenOParen:
		processMatching(hclsyntax.TokenCParen)
	case hclsyntax.TokenCQuote,
		hclsyntax.TokenCBrack,
		hclsyntax.TokenCBrace,
		hclsyntax.TokenCHeredoc,
		hclsyntax.TokenCParen:
		r.depth--
		r.matching[token.Type]--
	}

	if token.Type == hclsyntax.TokenIdent && !r.inside && r.depth == 1 {
		name := string(token.Bytes)
		if name == r.Name {
			r.inside = true
			r.waitFor = hclsyntax.TokenEqual
		}
		return Tokens{&token}, false
	} else if r.inside {
		// we're currently inside the value we need to cut
		// and we found the next token we're looking for
		if token.Type == r.waitFor {
			switch token.Type {
			case hclsyntax.TokenEqual:
				// if it's '=' just, skip it and set new waitFor token
				r.waitFor = hclsyntax.TokenEOF // we don't know yet
				return Tokens{&token}, false
			case hclsyntax.TokenNewline:
				// we finished, let's replace value and end the job
				r.inside = false
				tokens := r.Value
				tokens = append(tokens, &token)
				return tokens, false
			case hclsyntax.TokenCBrack, hclsyntax.TokenCQuote, hclsyntax.TokenCBrace, hclsyntax.TokenCHeredoc, hclsyntax.TokenCParen:
				if r.matching[token.Type] > 0 {
					// this is internal token, we not finished, skip it
					return Tokens{&token}, true
				}
				r.waitFor = hclsyntax.TokenNewline // just wait till final newline
				return Tokens{&token}, true
			}
		}

	}

	return Tokens{&token}, r.inside
}
