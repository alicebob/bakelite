package bakelite

import (
	"strings"
	"testing"

	"github.com/alicebob/bakelite/internal"
)

func TestTableLeaf(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		page := make([]byte, pageSize)
		ok(t, makeTableLeaf(page, false, nil))

		// and check our work
		tree, err := internal.NewBtree(page, false, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 0, len(leaf.Cells))
	})

	t.Run("single row", func(t *testing.T) {
		cell := tableLeafCell{
			left:    42,
			payload: []byte("hello world"),
		}
		page := make([]byte, pageSize)
		ok(t, makeTableLeaf(page, false, []tableLeafCell{cell}))

		// and check our work
		tree, err := internal.NewBtree(page, false, pageSize)
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
		page := make([]byte, pageSize)
		ok(t, makeTableLeaf(page, true, []tableLeafCell{cell}))

		// and check our work
		tree, err := internal.NewBtree(page, true, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 1, len(leaf.Cells))
		eq(t, cell.payload, leaf.Cells[0].Payload.Payload)
	})

	t.Run("bunch of rows", func(t *testing.T) {
		payload := []byte(strings.Repeat("hello world", 4))
		var cells []tableLeafCell
		for i := 0; i < 10; i++ {
			cells = append(cells, tableLeafCell{
				left:     i,
				fullSize: len(payload),
				payload:  payload,
			})
		}
		page := make([]byte, pageSize)
		ok(t, makeTableLeaf(page, false, cells))

		// and check our work
		tree, err := internal.NewBtree(page, false, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 10, len(leaf.Cells))
		eq(t, int64(0), leaf.Cells[0].Left)
		eq(t, payload, leaf.Cells[0].Payload.Payload)
	})
}
