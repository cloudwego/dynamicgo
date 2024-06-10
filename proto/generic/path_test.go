package generic

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTreeMarshal(t *testing.T) {
	desc := getExample3Desc()
	data := getExample3Data()
	partdesc := getExamplePartialDesc()
	t.Run("marshalNormal", func(t *testing.T) {
		v := NewRootValue(desc, data)
		tree := PathNode{
			Node: v.Node,
		}
		tree.Load(true, opts, desc)
		buf, err := tree.Marshal(opts)
		require.Nil(t, err)
		require.Equal(t, len(buf), len(data))
	})

	t.Run("marshalwithunknown", func(t *testing.T) {
		v := NewRootValue(partdesc, data)
		tree := PathNode{
			Node: v.Node,
		}
		tree.Load(true, opts, partdesc)
		buf, err := tree.Marshal(opts)
		require.Nil(t, err)
		require.Equal(t, len(buf), len(data))

	})
}
