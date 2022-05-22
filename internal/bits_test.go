package internal

import (
	"testing"
)

func TestVarint(t *testing.T) {
	test := func(n int64) {
		t.Helper()

		b := make([]byte, 100)
		enc := PutUvarint(b, uint64(n))

		t.Logf("test: %d -> %d bytes: %32b", n, enc, b[:enc])

		n2, l := ReadVarint(b[:enc])

		if have, want := l, enc; have != want {
			t.Errorf("bytes: have %d, want %d", have, want)
		}
		if have, want := n2, n; have != want {
			t.Errorf("value: have %d, want %d", have, want)
		}
	}

	test(0)
	test(1)
	test(2)
	test(127)
	test(128)
	test(129)
	test(240)
	test(241)
	test(16256)
	test(0x3FFF)
	test(0x7FFFFFFFFFFFFFFF)
	test(-1)
	test(-1000000)
}

func BenchmarkVarint(b *testing.B) {
	buf := make([]byte, 10)
	for i := 0; i < b.N; i++ {
		enc := PutUvarint(buf, uint64(12345999999))
		if have, want := 5, enc; have != want {
			b.Errorf("bytes: have %d, want %d", have, want)
		}
	}
}
