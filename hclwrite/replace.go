package hclwrite

// ReplaceFunc defines a token that is responsible for implementation
// of the token replacement logic. It's a helper function, and is
// provided from the caller because we try to minimize the changes to
// the hcl library.
type Replacer interface {
	ProcessToken(Token) (Tokens, bool) // consume the token and return replaced or original ones
	IsInside() bool                    // are we inside the value being replaced
}

func (e *Expression) ReplaceObjectField(replace Replacer) {
	ii := iterator{
		r: replace,
	}
	e.children = ii.iterateNodes(e.children)
}

type iterator struct {
	r Replacer
}

func (ii *iterator) iterateNodes(ns *nodes) *nodes {
	newNodes := &nodes{}
	for n := ns.first; n != nil; n = n.after {
		if tokens, ok := n.content.(Tokens); ok {
			newTokens := Tokens{}
			for _, token := range tokens {
				tt, skipped := ii.r.ProcessToken(*token)
				if skipped {
					continue
				}
				newTokens = append(newTokens, tt...)
			}
			n.content = newTokens
			newNodes.AppendNode(n)
		} else {
			if !ii.r.IsInside() {
				newNodes.AppendNode(n)
			}
		}
	}

	return newNodes
}
