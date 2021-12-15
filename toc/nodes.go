package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type nodes []*node

func (n nodes) Less(i, j int) bool {
	return n[i].weight < n[j].weight
}

func (n nodes) Len() int {
	return len(n)
}

func (n nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n nodes) find(s string) (*node, bool) {
	for _, e := range n {
		if v, ok := e.find(s); ok {
			return v, true
		}
	}
	return nil, false
}

func (n nodes) String() string {
	sort.Sort(n)

	sb := &strings.Builder{}
	fmt.Fprintf(sb, "# Summary\n")
	fmt.Fprintf(sb, "---\n")
	fmt.Fprintf(sb, "headless: true\n")
	fmt.Fprintf(sb, "bookhidden: true\n")
	fmt.Fprintf(sb, "---\n\n")

	for _, e := range n {
		fmt.Fprintf(sb, "%s", e)
		//fmt.Fprintf(sb, "%s%s\n", strings.Repeat("  ", e.indent), e.name)
	}
	return sb.String()
}

func (n *node) find(path string) (*node, bool) {
	if n.path == path {
		return n, true
	}
	for _, e := range n.subnodes {
		if v, ok := e.find(path); ok {
			return v, ok
		}
	}
	return nil, false
}

func (n *node) addSubNode(m *node) {
	m.indent = n.indent + 1
	n.subnodes = append(n.subnodes, m)
}

func (n *node) String() string {
	sb := &strings.Builder{}
	sort.Sort(n.subnodes)

	// write current node first
	prefix := fmt.Sprintf("%s%c", dir, filepath.Separator)
	rel := strings.TrimPrefix(n.path, prefix)
	fmt.Fprintf(sb, "%s* [%s](%s)\n", strings.Repeat("  ", n.indent), n.name, rel)

	// write subnodes if it has
	for _, e := range n.subnodes {
		fmt.Fprintf(sb, "%s", e.String())
	}
	return sb.String()
}
