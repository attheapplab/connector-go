package connector

import (
	"testing"
)

func TestTraverseRoot(t *testing.T) {
	n := &node{
		name: "",
	}

	path := "/"
	segments := pathToSegments(path)
	foundNode, _ := n.traverse(segments)

	want := ""
	if foundNode.name != "" {
		t.Errorf("\nhave: %s\nwant: %s", foundNode.name, want)
	}
}

func TestTraverseFooBarBaz(t *testing.T) {
		abc := &node{
			name: "abc",
			children: []*node{},
		}
				baz := &node{
					name: "baz",
				}
			bar := &node{
				name: "bar",
				children: []*node{baz},
			}
		foo := &node{
			name: "foo",
			children: []*node{bar},
		}
	root := &node{
		name: "",
		children: []*node{abc, foo},
	}

	path := "/foo/bar/baz"
	segments := pathToSegments(path)
	foundNode, _ := root.traverse(segments)
	
	want := "baz"
	if foundNode.name != want {
		t.Errorf("\nhave: %s\nwant: %s", foundNode.name, want)
	}
}

func TestTraverseWithIdentifier(t *testing.T) {
		abc := &node{
			name: "abc",
			children: []*node{},
		}
				baz := &node{
					name: "baz",
				}
			bar := &node{
				name: "bar",
				children: []*node{baz},
			}
		foo := &node{
			isIdentifier: true,
			name: "{foo}",
			children: []*node{bar},
		}
	root := &node{
		name: "",
		children: []*node{abc, foo},
	}

	path := "/woof/bar/baz"
	segments := pathToSegments(path)
	foundNode, _ := root.traverse(segments)
	
	want := "baz"
	if foundNode.name != want {
		t.Errorf("\nhave: %s\nwant: %s", foundNode.name, want)
	}
}

func TestAddAbc(t *testing.T) {
	root := &node{
		name: "",
	}

	path := "/foo/bar/baz/abc"
	segments := pathToSegments(path)
	foundNode, next := root.traverse(segments)
	foundNode.add(next)
	foundNode, next = root.traverse(segments)

	want := "abc"
	if foundNode.name != "abc" {
		t.Errorf("\nhave: %s\nwant: %s", foundNode.name, want)
	}
}
