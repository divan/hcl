package hclwrite

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Replacer implements token-level value replacer.
type Replacer struct {
	Name  string
	Value Tokens

	inside bool
}

// NewReplacer returns and inits new Replacer.
func NewReplacer(name string, value Tokens) *Replacer {
	return &Replacer{
		Name:  name,
		Value: value,
	}
}

func (r *Replacer) processToken(token Token, depth string) (Tokens, bool) {
	if token.Type == hclsyntax.TokenIdent && !r.inside {
		name := string(token.Bytes)
		if name == r.Name {
			r.inside = true
		}
		return Tokens{&token}, false
	} else if token.Type == hclsyntax.TokenEqual && r.inside {
		return Tokens{&token}, false
	} else if token.Type == hclsyntax.TokenNewline && r.inside {
		r.inside = false
		tokens := r.Value
		tokens = append(tokens, &token)
		return tokens, false
	}

	return Tokens{&token}, r.inside
}
