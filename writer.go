package rotate

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

type Writer struct {
	path    string
	f       *os.File
	size    int64
	maxSize int64
	keepGen int
	m       sync.Mutex
}

func NewWriter(path string, maxSize int64, keepGen int) *Writer {
	return &Writer{
		path:    path,
		maxSize: maxSize,
		keepGen: keepGen,
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.m.Lock()
	defer w.m.Unlock()
	if w.f == nil {
		if err = w.open(); err != nil {
			return
		}
	}
	if n, err = w.f.Write(p); err != nil {
		return
	}
	w.size += int64(n)
	if w.size >= w.maxSize {
		err = w.rotate()
	}
	return
}

func (w *Writer) open() error {
	f, err := os.OpenFile(w.path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	w.f, w.size = f, fi.Size()
	return nil
}

func (w *Writer) rotate() error {
	if err := w.f.Close(); err != nil {
		return err
	}
	w.f = nil
	if err := w.clean(); err != nil {
		return err
	}
	return os.Rename(w.path, w.path+".0")
}

func (w *Writer) clean() error {
	files, err := filepath.Glob(w.path + ".*")
	if err != nil {
		return err
	}
	sort.Sort(byGen(files))
	if len(files) >= w.keepGen {
		files = files[:w.keepGen-1]
	}
	for i := len(files) - 1; i >= 0; i-- {
		n, _ := strconv.Atoi(files[i][len(w.path)+1:])
		if err := os.Rename(files[i], w.path+"."+strconv.Itoa(n+1)); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) Close() error {
	w.m.Lock()
	defer w.m.Unlock()
	if w.f != nil {
		if err := w.f.Close(); err != nil {
			return err
		}
		w.f = nil
	}
	return nil
}

type byGen []string

func (z byGen) Len() int      { return len(z) }
func (z byGen) Swap(i, j int) { z[i], z[j] = z[j], z[i] }
func (z byGen) Less(i, j int) bool {
	a, _ := strconv.Atoi(filepath.Ext(z[i]))
	b, _ := strconv.Atoi(filepath.Ext(z[j]))
	return a < b
}
