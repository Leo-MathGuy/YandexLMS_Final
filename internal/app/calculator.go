package app

import (
	"fmt"
	"strconv"
	"strings"
)

type Node struct {
	isValue bool
	value   *float64
	left    *Node
	right   *Node
	op      *rune
	parent  *Node
}

func (n *Node) toString() string {
	if n.isValue {
		return strconv.FormatFloat(*n.value, 'f', -1, 64)
	} else {
		return fmt.Sprintf("[%s%s%s]", n.left.toString(), string(*n.op), n.right.toString())
	}
}

const (
	expr = iota
	term
	factor
)

type nodeproc func([]ExprToken, uint, nodeproc, int) *Node

// Generate AST
func NodeGen(tokens []ExprToken, mode uint, f nodeproc, level int) *Node {
	fmt.Print("Entering on level " + fmt.Sprint(level))
	if len(tokens) == 1 {
		fmt.Println(" returning single value " + fmt.Sprint(*tokens[0].valueF))
		return &Node{true, tokens[0].valueF, nil, nil, nil, nil}
	}

	sep := ""
	root := Node{}

	switch mode {
	case expr:
		fmt.Println(" mode +-")
		sep = "+-"
	case term:
		fmt.Println(" mode */")
		sep = "*/"
	case factor:
		// THIS SHOULD NEVER HAPPEN
		fmt.Println(" should not happen factor:")
		for _, v := range tokens {
			if v.tokenType == number {
				fmt.Printf(" - Number - %f\n", *v.valueF)
			} else {
				fmt.Printf(" - %d - %d\n", v.tokenType, *v.valueI)
			}
		}

		// To guarantee segfault
		return nil
	}

	curPart := make([]ExprToken, 0)
	current := &root
	skip := 0
	parCount := 0
	next := false
	skipGen := false

	for i, v := range tokens {
		fmt.Printf("%d - ", i)

		if skip > 0 {
			skip--
			continue
		}

		switch v.tokenType {
		case operator:
			r := rune(*v.valueI)
			if strings.Contains(sep, string(r)) && parCount == 0 {
				if !skipGen {
					current.left = f(curPart, mode+1, f, level+1)
				} else {
					skipGen = false
				}
				current.isValue = false
				current.op = &r
				current.right = &Node{}
				current.right.parent = current
				current = current.right
				curPart = make([]ExprToken, 0)
				next = true
				fmt.Println("Op " + string(rune(r)))
				continue
			} else {
				fmt.Println("Adding an op " + string(rune(*v.valueI)))
			}
		case parentheses:
			if next {
				subPar := 0
				for j, x := range tokens[i+1:] {
					skip++
					if x.tokenType == parentheses {
						if *x.valueI == int(')') {
							if subPar == 0 {
								fmt.Println("Entering parentheses")
								generated := f(tokens[i+1:i+skip], expr, f, level+1)
								for _, v := range tokens[i+1 : i+skip] {
									if v.tokenType == number {
										fmt.Printf("Added a %f\n", *v.valueF)
									} else {
										fmt.Printf("Added a %s\n", string(rune(*v.valueI)))
									}
								}
								fmt.Println("Exiting parentheses")
								if j+i == len(tokens)-2 {
									current.parent.right = generated
								} else {
									current.left = generated
								}
								skipGen = true
							} else {
								subPar--
							}
						} else {
							subPar++
						}
					}
				}
				continue
			} else {
				switch rune(*v.valueI) {
				case '(':
					parCount++
				case ')':
					parCount--
				}
				fmt.Println("Adding a par " + string(rune(*v.valueI)))
			}
		default:
			if v.tokenType == number {
				fmt.Println("Adding a number " + fmt.Sprint(*v.valueF))
			} else {
				fmt.Println("Adding a " + string(rune(*v.valueI)))
			}
			next = false
		}
		curPart = append(curPart, v)
	}

	if root.left != nil {
		if !skipGen {
			current.parent.right = f(curPart, mode+1, f, level+1)
		}
		return &root
	} else {
		return f(curPart, mode+1, f, level+1)
	}
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
			fmt.Printf("Detected a unary - at %d\n", i)
			if tokens[i+1].tokenType == number {
				fmt.Println(" - inverse")
				*tokens[i+1].valueF *= -1
			} else {
				fmt.Println(" - -1 *")
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
