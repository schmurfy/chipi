package gen

import (
	"testing"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComments(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Comments parser", func() {
		g.It("should parse multi-line comments", func() {
			lines := []string{
				"// @description",
				"// this is a wonderful path with",
				"// a lot of things inside, really a great path !",
			}

			ret, err := parseComment(lines)
			require.NoError(g, err)

			assert.Equal(g, "this is a wonderful path with\na lot of things inside, really a great path !\n", ret["description"])
		})

		g.It("should parse block comments")
	})
}
