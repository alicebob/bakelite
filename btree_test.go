package bakelite

import (
	"strings"
	"testing"

	"github.com/alicebob/bakelite/internal"
)

func TestTableLeaf(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		page := make([]byte, PageSize)
		writeTableLeaf(page, false, nil)

		// and check our work
		tree, err := internal.NewBtree(page, false, PageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 0, len(leaf.Cells))
	})

	t.Run("single row", func(t *testing.T) {
		payload := []byte("hello world")
		cell := leafCell(42, len(payload), payload, 0)
		page := make([]byte, PageSize)
		writeTableLeaf(page, false, []tableLeafCell{cell})

		// and check our work
		tree, err := internal.NewBtree(page, false, PageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 1, len(leaf.Cells))
		eq(t, payload, leaf.Cells[0].Payload.Payload)
	})

	t.Run("special page 1", func(t *testing.T) {
		payload := []byte("hello world")
		cell := leafCell(42, len(payload), payload, 0)
		page := make([]byte, PageSize)
		writeTableLeaf(page, true, []tableLeafCell{cell})

		// and check our work
		tree, err := internal.NewBtree(page, true, PageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 1, len(leaf.Cells))
		eq(t, payload, leaf.Cells[0].Payload.Payload)
	})

	t.Run("bunch of rows", func(t *testing.T) {
		payload := []byte(strings.Repeat("hello world", 4))
		var cells []tableLeafCell
		for i := 0; i < 10; i++ {
			cells = append(cells, leafCell(i, len(payload), payload, 0))
		}
		page := make([]byte, PageSize)
		writeTableLeaf(page, false, cells)

		// and check our work
		tree, err := internal.NewBtree(page, false, PageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 10, len(leaf.Cells))
		eq(t, int64(0), leaf.Cells[0].Left)
		eq(t, payload, leaf.Cells[0].Payload.Payload)
	})

	t.Run("many rows", func(t *testing.T) {
		payload := []byte(strings.Repeat("helloworld", 100)) // 1K
		var cells []tableLeafCell
		for i := 0; i < 4; i++ {
			cells = append(cells, leafCell(i, len(payload), payload, 0))
		}
		page := make([]byte, PageSize)
		writeTableLeaf(page, false, cells)

		// and check our work
		tree, err := internal.NewBtree(page, false, PageSize)
		ok(t, err)
		leaf := tree.(*internal.TableLeaf)
		eq(t, 4, len(leaf.Cells))
		eq(t, int64(0), leaf.Cells[0].Left)
		eq(t, payload, leaf.Cells[0].Payload.Payload)
	})
}

func TestTableInterior(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		page := make([]byte, PageSize)
		cells := []tableInteriorCell{
			{key: 42, left: 12},
		}
		eq(t, 1, writeTableInterior(page, false, cells))

		// and check our work
		tree, err := internal.NewBtree(page, false, PageSize)
		ok(t, err)
		leaf := tree.(*internal.TableInterior)
		eq(t, 0, len(leaf.Cells))
		eq(t, 12, leaf.Rightmost)
	})

	t.Run("two", func(t *testing.T) {
		page := make([]byte, PageSize)
		cells := []tableInteriorCell{
			{key: 42, left: 12},
			{key: 84, left: 13},
		}
		eq(t, 2, writeTableInterior(page, false, cells))

		// and check our work
		tree, err := internal.NewBtree(page, false, PageSize)
		ok(t, err)
		leaf := tree.(*internal.TableInterior)
		eq(t, 13, leaf.Rightmost)
		eq(t, 1, len(leaf.Cells))
		eq(t, internal.TableInteriorCell{Left: 12, Key: 84}, leaf.Cells[0])
	})

	t.Run("three", func(t *testing.T) {
		page := make([]byte, PageSize)
		cells := []tableInteriorCell{
			{key: 42, left: 12},
			{key: 84, left: 13},
			{key: 102, left: 9},
		}
		eq(t, 3, writeTableInterior(page, false, cells))

		// and check our work
		tree, err := internal.NewBtree(page, false, PageSize)
		ok(t, err)
		leaf := tree.(*internal.TableInterior)
		eq(t, 9, leaf.Rightmost)
		eq(t, 2, len(leaf.Cells))
		eq(t, internal.TableInteriorCell{Left: 12, Key: 84}, leaf.Cells[0])
		eq(t, internal.TableInteriorCell{Left: 13, Key: 102}, leaf.Cells[1])
	})
}
