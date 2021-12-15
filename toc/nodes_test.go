package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var txt = `# Summary
---
headless: true
bookhidden: true
---

* [1](1)
  * [1.1](1.1)
  * [1.2](1.2)
* [2](2)
  * [2.1](2.1)
  * [2.2](2.2)
`
var all = nodes{
	&node{
		name:   "1",
		path:   "1",
		weight: 1,
		indent: 0,
		subnodes: nodes{
			&node{
				name:   "1.1",
				path:   "1.1",
				weight: 1,
				indent: 1,
			},
			&node{
				name:   "1.2",
				path:   "1.2",
				weight: 2,
				indent: 1,
			},
		},
	},
	&node{
		name:   "2",
		path:   "2",
		weight: 1,
		indent: 0,
		subnodes: nodes{
			&node{
				name:   "2.1",
				path:   "2.1",
				weight: 1,
				indent: 1,
			},
			&node{
				name:   "2.2",
				path:   "2.2",
				weight: 2,
				indent: 1,
			},
		},
	},
}

func TestNodes_String(t *testing.T) {
	assert.Equal(t, all.String(), txt)
}

func TestNodes_find(t *testing.T) {
	v, ok := all.find("1.1")
	assert.True(t, ok)
	assert.Equal(t, v.name, "1.1")
}

func TestNodes_add(t *testing.T) {
	v, ok := all.find("1.1")
	assert.True(t, ok)

	v.addSubNode(&node{
		name:   "1.1.1",
		path:   "1.1.1",
		weight: 1,
	})

	assert.Equal(t, 1, v.subnodes.Len())
}
