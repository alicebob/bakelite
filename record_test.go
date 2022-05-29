package bakelite

import (
	"fmt"
	"testing"

	"github.com/alicebob/bakelite/internal"
)

func TestRecord(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		t.Skip("not sure this is valid")
		bs, err := makeRecord(nil)
		ok(t, err)

		fmt.Printf("bs: %q\n", bs)
		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 0, len(rec))
	})

	t.Run("8-bit int", func(t *testing.T) {
		bs, err := makeRecord([]any{42})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 1, len(rec))
		eq(t, 42, rec[0].(int64))
	})

	t.Run("16-bit int", func(t *testing.T) {
		bs, err := makeRecord([]any{1 << 14})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 1, len(rec))
		eq(t, 1<<14, rec[0].(int64))
	})

	t.Run("24-bit int", func(t *testing.T) {
		bs, err := makeRecord([]any{1 << 20})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 1, len(rec))
		eq(t, 1<<20, rec[0].(int64))
	})

	t.Run("32-bit int", func(t *testing.T) {
		bs, err := makeRecord([]any{1 << 30})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 1, len(rec))
		eq(t, 1<<30, rec[0].(int64))
	})

	t.Run("48-bit int", func(t *testing.T) {
		bs, err := makeRecord([]any{1 << 45})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 1, len(rec))
		eq(t, 1<<45, rec[0].(int64))
	})

	t.Run("64-bit int", func(t *testing.T) {
		bs, err := makeRecord([]any{1 << 62})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 1, len(rec))
		eq(t, 1<<62, rec[0].(int64))
	})

	t.Run("simple string", func(t *testing.T) {
		bs, err := makeRecord([]any{"hello"})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 1, len(rec))
		eq(t, "hello", rec[0].(string))
	})

	t.Run("special case: 0", func(t *testing.T) {
		bs, err := makeRecord([]any{0})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 1, len(rec))
		eq(t, 0, rec[0].(int64))
	})

	t.Run("special case: 1", func(t *testing.T) {
		bs, err := makeRecord([]any{1})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 1, len(rec))
		eq(t, 1, rec[0].(int64))
	})

	t.Run("null", func(t *testing.T) {
		bs, err := makeRecord([]any{nil})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 1, len(rec))
		eq(t, nil, rec[0])
	})

	t.Run("bunch", func(t *testing.T) {
		bs, err := makeRecord([]any{1, "hello", -45, nil, 4})
		ok(t, err)

		rec, err := internal.ParseRecord(bs)
		ok(t, err)
		eq(t, 5, len(rec))
		eq(t, 1, rec[0].(int64))
		eq(t, "hello", rec[1].(string))
		eq(t, -45, rec[2].(int64))
		eq(t, nil, rec[3])
		eq(t, 4, rec[4].(int64))
	})
}
