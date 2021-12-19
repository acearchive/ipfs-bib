package archive

import (
	"golang.org/x/net/html"
)

func FindAttr(node *html.Node, key string) (value *string) {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return &attr.Val
		}
	}

	return nil
}

func FindChild(node *html.Node, predicate func(*html.Node) bool) *html.Node {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if predicate(child) {
			return child
		}
	}

	return nil
}

type TagFinder struct {
	predicate func(*html.Node) bool
}

func NewTagFinder(predicate func(*html.Node) bool) *TagFinder {
	return &TagFinder{predicate}
}

func findSiblings(node *html.Node) []html.Node {
	var siblings []html.Node

	if node.PrevSibling != nil {
		panic("this method only accepts the first sibling in the node")
	}

	for sibling := node; sibling != nil; sibling = sibling.NextSibling {
		siblings = append(siblings, *sibling)
	}

	return siblings
}

func (f *TagFinder) walk(nodes []html.Node) *html.Node {
	for _, node := range nodes {
		if f.predicate(&node) {
			return &node
		}
	}

	for _, node := range nodes {
		if child := node.FirstChild; child != nil {
			if result := f.walk(findSiblings(child)); result != nil {
				return result
			}
		}
	}

	return nil
}

func (f *TagFinder) Find(node *html.Node) *html.Node {
	if child := node.FirstChild; child != nil {
		return f.walk(findSiblings(child))
	}

	return nil
}
