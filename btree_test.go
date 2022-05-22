package bakelite

import (
	"strings"
	"testing"

	"github.com/alicebob/bakelite/internal"
)

func TestTableLeaf(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		page := make([]byte, pageSize)
		eq(t, 0, writeTableLeaf(page, false, nil))

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
		eq(t, 1, writeTableLeaf(page, false, []tableLeafCell{cell}))

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
		eq(t, 1, writeTableLeaf(page, true, []tableLeafCell{cell}))

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
		eq(t, 10, writeTableLeaf(page, false, cells))

		// and check our work
		tree, err := internal.NewBtree(page, false, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 10, len(leaf.Cells))
		eq(t, int64(0), leaf.Cells[0].Left)
		eq(t, payload, leaf.Cells[0].Payload.Payload)
	})

	t.Run("too many rows", func(t *testing.T) {
		payload := []byte(strings.Repeat("helloworld", 100)) // 1K
		var cells []tableLeafCell
		for i := 0; i < 10; i++ {
			cells = append(cells, tableLeafCell{
				left:     i,
				fullSize: len(payload),
				payload:  payload,
			})
		}
		expect := 4 // =~ page size / 1K
		page := make([]byte, pageSize)
		eq(t, expect, writeTableLeaf(page, false, cells))

		// and check our work
		tree, err := internal.NewBtree(page, false, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, expect, len(leaf.Cells))
		eq(t, int64(0), leaf.Cells[0].Left)
		eq(t, payload, leaf.Cells[0].Payload.Payload)
	})
}

func TestTableInterior(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		page := make([]byte, pageSize)
		cells := []tableInteriorCell{
			{key: 42, left: 12},
		}
		eq(t, 1, writeTableInterior(page, cells))

		// and check our work
		tree, err := internal.NewBtree(page, false, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableInterior)
		eq(t, 0, len(leaf.Cells))
		eq(t, 12, leaf.Rightmost)
	})

	t.Run("two", func(t *testing.T) {
		page := make([]byte, pageSize)
		cells := []tableInteriorCell{
			{key: 42, left: 12},
			{key: 84, left: 13},
		}
		eq(t, 2, writeTableInterior(page, cells))

		// and check our work
		tree, err := internal.NewBtree(page, false, pageSize)
		ok(t, err)
		leaf := tree.(*internal.TableInterior)
		eq(t, 13, leaf.Rightmost)
		eq(t, 1, len(leaf.Cells))
		eq(t, internal.TableInteriorCell{Left: 12, Key: 42}, leaf.Cells[0])
	})
}
