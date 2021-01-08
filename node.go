package connector

import (
	"regexp"
)

type node struct {
	children     []*node
	isIdentifier bool
	isMatch      bool
	name         string
	procedures   []procedure
}

func (n *node) traverse(segments []string) (*node, []string) {
	segment := segments[0]
	for _, child := range n.children {
		if segment == child.name || child.isIdentifier {
			next := segments[1:]
			child.isMatch = len(next) == 0
			if child.isMatch {
				return child, next
			}
			return child.traverse(next)
		}
	}
	return n, segments
}

func (n *node) add(segments []string, procedures []procedure) {
	segment := segments[0]
	isIdentifier, _ := regexp.MatchString("{([a-z]+)}", segment)
	newNode := &node{
		isIdentifier: isIdentifier,
		name:         segment,
	}
	n.children = append(n.children, newNode)
	next := segments[1:]
	if len(next) > 0 {
		newNode.add(next, procedures)
	} else {
		newNode.procedures = procedures
	}
}
