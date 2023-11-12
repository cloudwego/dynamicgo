package generic

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTreeMarshal(t *testing.T) {
	desc := getExample2Desc()
	data := getExample2Data()
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

	
	// opts := Options{}
	// err := tree.Assgin(true, &opts)
	// require.NoError(t, err)

	// out, err := tree.Marshal(&opts)
	// require.Nil(t, err)
	// spew.Dump(out)
	// exp := example2.NewExampleReq()
	// _, err = exp.FastRead(out)
	// require.Nil(t, err)

	// x := v.GetByPath(PathExampleByte...)
	// tt := PathNode{
	// 	Path: NewPathFieldName("Msg"),
	// 	Node: x.Node,
	// 	Next: []PathNode{
	// 		tree,
	// 	},
	// }
	// out, err = tt.Marshal(&opts)
	// require.Nil(t, err)
	// require.Equal(t, x.Raw(), out)
}