package bakelite

import (
	"strings"
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

	t.Run("single row", func(t *testing.T) {
		cell := tableLeafCell{
			left:    42,
			payload: []byte("hello world"),
		}
		bs, err := makeTableLeaf(false, []tableLeafCell{cell})
		ok(t, err)

		tree, err := internal.NewBtree(bs, false, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 1, len(leaf.Cells))
		eq(t, cell.payload, leaf.Cells[0].Payload.Payload)
	})

	t.Run("special page 1", func(t *testing.T) {
		cell := tableLeafCell{
			left:    42,
			payload: []byte("hello world"),
		}
		bs, err := makeTableLeaf(true, []tableLeafCell{cell})
		ok(t, err)

		tree, err := internal.NewBtree(bs, true, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 1, len(leaf.Cells))
		eq(t, cell.payload, leaf.Cells[0].Payload.Payload)
	})

	t.Run("bunch of rows", func(t *testing.T) {
		payload := []byte(strings.Repeat("hello world", 40))
		var cells []tableLeafCell
		for i := 0; i < 10; i++ {
			cells = append(cells, tableLeafCell{
				left:    int64(i),
				payload: payload,
			})
		}
		bs, err := makeTableLeaf(false, cells)
		ok(t, err)

		tree, err := internal.NewBtree(bs, false, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 10, len(leaf.Cells))
		eq(t, int64(0), leaf.Cells[0].Left)
		eq(t, payload, leaf.Cells[0].Payload.Payload)
	})
}
