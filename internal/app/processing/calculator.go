package processing

import (
	"fmt"
	"strconv"
	"strings"
)

// The node of the AST
type Node struct {
	IsValue bool
	Value   *float64
	Left    *Node
	Right   *Node
	Op      *rune
	Parent  *Node
}

// For testing purposes
func (n *Node) toString() string {
	if n.IsValue {
		return strconv.FormatFloat(*n.Value, 'f', -1, 64)
	} else {
		return fmt.Sprintf("[%s%s%s]", n.Left.toString(), string(*n.Op), n.Right.toString())
	}
}

// Levels of the Recursive Descent Parsing algorithm
const (
	expr = iota
	term
	factor
)

// Function to test the NodeGen function specifically
type nodeproc func([]ExprToken, uint, nodeproc, int) *Node

// Recursive Descent Parsing algorithm
// CANNOT HANDLE UNARY
func NodeGen(tokens []ExprToken, mode uint, f nodeproc, level int) *Node {
	// Shortcut for single numbers
	if len(tokens) == 1 {
		return &Node{IsValue: true, Value: tokens[0].valueF}
	}

	// For expr and term we split on current-level operators,
	// but build the tree left-associatively.
	if mode == expr || mode == term {
		sep := "+-"
		if mode == term {
			sep = "*/"
		}

		var tree *Node // root of our growing AST
		var last *Node // the most recent operator node
		var currentExpr []ExprToken
		paren := 0

		for _, tok := range tokens {
			if tok.tokenType == parentheses {
				switch rune(*tok.valueI) {
				case '(':
					paren++
				case ')':
					paren--
				}
			}

			if tok.tokenType == operator && paren == 0 && strings.ContainsRune(sep, rune(*tok.valueI)) {
				// first, if this isn’t the very first operator, close out the previous one:
				if last != nil {
					last.Right = f(currentExpr, mode+1, f, level+1)
					last.Right.Parent = last
				}

				// build the new operator node
				opNode := &Node{IsValue: false, Op: RunePtr(rune(*tok.valueI))}
				if tree == nil {
					// very first operator: its left child is everything so far
					opNode.Left = f(currentExpr, mode+1, f, level+1)
					opNode.Left.Parent = opNode
					tree = opNode
				} else {
					// subsequent operator: make it the new root,
					// with left = entire old tree
					opNode.Left = tree
					tree.Parent = opNode
					tree = opNode
				}
				// prepare its right child placeholder:
				opNode.Right = &Node{}
				opNode.Right.Parent = opNode

				// and remember this as the last operator we need to close
				last = opNode

				// reset tokens accumulator
				currentExpr = nil
			} else {
				currentExpr = append(currentExpr, tok)
			}
		}

		if last != nil {
			// close out the final operator:
			last.Right = f(currentExpr, mode+1, f, level+1)
			last.Right.Parent = last
			return tree
		}
		// no top-level operator found → dive one level deeper
		return f(tokens, mode+1, f, level+1)
	}

	// finally, factor must be a parenthesized sub-expr
	if mode == factor {
		return f(tokens[1:len(tokens)-1], expr, f, level+1)
	}

	return nil
}

// Assumes expression has passed validation
func Eval(tokens []ExprToken, f nodeproc) (*Node, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("you gotta be kidding me")
	}

	// Process unary - and negative numbers
	processed := make([]ExprToken, 0)

	for i, v := range tokens {
		if v.tokenType == operator && *v.valueI == int('-') && (i == 0 || tokens[i-1].tokenType == operator || (tokens[i-1].tokenType == parentheses && *tokens[i-1].valueI == int('('))) {
			if tokens[i+1].tokenType == number {
				*tokens[i+1].valueF *= -1
			} else {
				processed = append(
					processed,
					ExprToken{number, FloatPtr(-1.0), nil},
					ExprToken{operator, nil, IntPtr(int('*'))},
				)
			}
		} else {
			processed = append(processed, v)
		}
	}

	return NodeGen(processed, expr, f, 0), nil
}
