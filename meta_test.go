package bakelite

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// ok fails the test if err is not nil.
func ok(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %s", err.Error())
	}
}

// compare two values
func eq[A any](tb testing.TB, want, have A) {
	tb.Helper()
	if !reflect.DeepEqual(have, want) {
		s := fmt.Sprintf("%v", want)
		if !strings.ContainsAny(s, "\n\r\t") && len(s) < 30 {
			tb.Fatalf("have %v, want %v", have, want)
		}

		tb.Fatalf("equal error:\n - have:\n%#v - want:\n%#v", have, want)
		// tb.Fatalf("equal error:\n - have:\n%s - want:\n%s", spew.Sdump(have), spew.Sdump(want))
	}
}

type Tuple[A any] struct {
	v   A
	err error
}

func tuple[A any](v A, err error) Tuple[A] {
	return Tuple[A]{
		v:   v,
		err: err,
	}
}

// use: mustEq(t, "hello", func() (string, error) { return "hello", nil})
func mustEq[A any](tb testing.TB, want A, have Tuple[A]) {
	tb.Helper()
	ok(tb, have.err)
	eq(tb, have.v, want)
}

// save file under ./testdata/<name>
// They shouldn't be checked in, but we keep them around for easier manual checks
func saveFile(t *testing.T, b *bytes.Buffer, name string) string {
	t.Helper()
	file := "./testdata/" + name
	ok(t, os.WriteFile(file, b.Bytes(), 0666))
	return file
}
