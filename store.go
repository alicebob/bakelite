package bakelite

import (
	"io"
	"os"
)

// Stores pages
type storer interface {
	// Add a new page, returns the ID
	AddPage(page []byte) (int, error)

	WriteTo(page1 []byte, w io.Writer) error

	// Close with whatever cleanup
	Close() error
}

type memStore struct {
	pages [][]byte
}

func (m *memStore) AddPage(page []byte) (int, error) {
	m.pages = append(m.pages, page)
	return len(m.pages), nil
}

func (m *memStore) WriteTo(page1 []byte, w io.Writer) error {
	if _, err := w.Write(page1); err != nil {
		return err
	}
	for _, p := range m.pages[1:] {
		if _, err := w.Write(p); err != nil {
			return err
		}
	}
	return nil
}

func (m *memStore) Close() error {
	m.pages = nil
	return nil
}

type fileStore struct {
	f      *os.File
	pageID int // 0 based
}

func newTmpStore(dir string) (*fileStore, error) {
	f, err := os.CreateTemp(dir, "bakelite-*") // these aren't full sqlite files
	if err != nil {
		return nil, err
	}

	return &fileStore{
		f:      f,
		pageID: 0,
	}, nil
}

func (f *fileStore) AddPage(page []byte) (int, error) {
	if _, err := f.f.Write(page); err != nil {
		return 0, err
	}
	f.pageID++
	return f.pageID, nil
}

func (f *fileStore) WriteTo(page1 []byte, w io.Writer) error {
	if _, err := w.Write(page1); err != nil {
		return err
	}
	if _, err := f.f.Seek(int64(len(page1)), 0); err != nil {
		return err
	}
	_, err := io.Copy(w, f.f)
	return err
}

func (f *fileStore) Close() error {
	name := f.f.Name()
	f.f.Close()
	return os.Remove(name)
}
