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
	// Shortcut for number values
	if len(tokens) == 1 {
		return &Node{true, tokens[0].valueF, nil, nil, nil, nil}
	}

	root := Node{}
	currentNode := &root
	currentExpr := make([]ExprToken, 0)
	paren := 0

	switch mode {
	case expr, term:
		sep := "*/" // What separates groups
		if mode == expr {
			sep = "+-"
		}

		for _, v := range tokens {
			// Treat expression in parentheses as one group
			if v.tokenType == parentheses {
				switch rune(*v.valueI) {
				case '(':
					paren++
				case ')':
					paren--
				}
			}

			if v.tokenType == operator && strings.Contains(sep, string(rune(*v.valueI))) && paren == 0 {
				// Advance the AST tree
				currentNode.Left = f(currentExpr, mode+1, f, level+1)
				currentNode.Right = &Node{}

				currentNode.Left.Parent = currentNode
				currentNode.Right.Parent = currentNode

				currentNode.IsValue = false
				currentNode.Op = RunePtr(rune(*v.valueI))
				currentNode = currentNode.Right
				currentExpr = make([]ExprToken, 0)
			} else {
				currentExpr = append(currentExpr, v)
			}
		}

		if currentNode.Parent != nil {
			// Finish the tree
			currentNode.Parent.Right = f(currentExpr, mode+1, f, level+1)
			return &root
		} else {
			// Only 1 group was found
			return f(currentExpr, mode+1, f, level+1)
		}

	case factor:
		// Parentheses expr
		return f(tokens[1:len(tokens)-1], expr, f, level+1)
	}

	panic("no") // 🗿
}

// Assumes expression has passed validation
func Eval(tokens []ExprToken, f nodeproc) (*Node, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("you gotta be kidding me")
	}

	// Process unary - and negative numbers
	processed := make([]ExprToken, 0)

	for i, v := range tokens {
		if v.tokenType == operator && *v.valueI == int('-') && (i == 0 || tokens[i-1].tokenType == operator) {
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
