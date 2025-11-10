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

func BenchmarkTreeUnmarshal(b *testing.B) {
	desc := getExample3Desc()
	data := getExample3Data()
	v := NewRootValue(desc, data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := NewPathNode()
		tree.Node = v.Node
		if err := tree.Load(true, opts, desc); err != nil {
			b.Fatal(err)
		}
		FreePathNode(tree)
	}
}

func BenchmarkTreeMarshal(b *testing.B) {
	desc := getExample3Desc()
	data := getExample3Data()
	v := NewRootValue(desc, data)
	tree := PathNode{
		Node: v.Node,
	}
	if err := tree.Load(true, opts, desc); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := tree.Marshal(opts)
		if err != nil {
			b.Fatal(err)
		}
		_ = buf
	}
}
