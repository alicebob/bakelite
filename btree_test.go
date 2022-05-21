package bakelite

import (
	"testing"

	"github.com/alicebob/bakelite/internal"
)

func TestTableLeaf(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		bs, err := makeTableLeaf(false, nil)
		ok(t, err)

		tree, err := internal.NewBtree(bs, false, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 0, len(leaf.Cells))
	})
}
