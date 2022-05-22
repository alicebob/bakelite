package internal

import (
	"testing"
)

func TestVarint(t *testing.T) {
	// encoded, bytes from the number, value
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

/*
func TestVarint(t *testing.T) {
	// encoded, bytes from the number, value
	test := func(eb []byte, el int, en int64) {
		t.Helper()
		t.Logf("test: %#08b -> %d? - %#032b", eb, en, en)
		n, l := ReadVarint(eb)
		if have, want := l, el; have != want {
			t.Errorf("read: have %d, want %d", have, want)
		}
		if have, want := n, en; have != want {
			t.Errorf("read: have %d, want %d", have, want)
		}

		if el < 0 || en < 0 {
			return
		}

		// ... and go back again
		b := make([]byte, 100)
		enc := PutUvarint(b, uint64(en))
		if have, want := enc, el; have != want {
			t.Errorf("put bytes: have %d, want %d", have, want)
		}
		if have := b[:enc]; !reflect.DeepEqual(have, eb) {
			t.Errorf("put: have %#08b, want %#08b", have, eb)
		}
	}
	test([]byte("\x00"), 1, 0)
	test([]byte("\x01"), 1, 1)
	test([]byte("\x02"), 1, 2)
	test([]byte{0b1000_0001, 0b0000_0000}, 2, 1<<8)
	test([]byte{0b1000_0001, 0b0000_0001}, 2, 1<<7+1)
	test([]byte{0b1000_0001, 0b0000_0010}, 2, 1<<7+2)
	test([]byte("\xFF\x00"), 2, 16256) // 0b00111111_10000000
	return
	test([]byte("\xFF\x7F"), 2, 0x3FFF) // 0b00111111_11111111
	test([]byte("\xBF\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF"), 9, 0x7FFFFFFFFFFFFFFF)
	return
	test([]byte("\xBF\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFFignored"), 9, 0x7FFFFFFFFFFFFFFF)
	// int64 overflow
	test([]byte("\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF"), 9, -1)
	// Error cases
	test([]byte("\xFF"), -1, 0)
	test([]byte("\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF"), -1, 0)
}
*/
