package app

import (
	"fmt"
	"strconv"
	"strings"
)

// The node of the AST
type Node struct {
	isValue bool
	value   *float64
	left    *Node
	right   *Node
	op      *rune
	parent  *Node
}

// For testing purposes
func (n *Node) toString() string {
	if n.isValue {
		return strconv.FormatFloat(*n.value, 'f', -1, 64)
	} else {
		return fmt.Sprintf("[%s%s%s]", n.left.toString(), string(*n.op), n.right.toString())
	}
}

// For testing purposes
func (n *Node) toStringShort() string {
	if n.isValue {
		return strconv.FormatFloat(*n.value, 'f', -1, 64)
	} else {
		return fmt.Sprintf("%s%s%s", n.left.toStringShort(), string(*n.op), n.right.toStringShort())
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
				currentNode.left = f(currentExpr, mode+1, f, level+1)
				currentNode.right = &Node{}

				currentNode.left.parent = currentNode
				currentNode.right.parent = currentNode

				currentNode.isValue = false
				currentNode.op = RunePtr(rune(*v.valueI))
				currentNode = currentNode.right
				currentExpr = make([]ExprToken, 0)
			} else {
				currentExpr = append(currentExpr, v)
			}
		}

		if currentNode.parent != nil {
			// Finish the tree
			currentNode.parent.right = f(currentExpr, mode+1, f, level+1)
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
