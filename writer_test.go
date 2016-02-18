package rotate

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriter(t *testing.T) {
	dir, err := ioutil.TempDir("", "rotate-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "test.log")
	w := NewWriter(path, 8, 2)
	write := func(s string) {
		_, err := w.Write([]byte(s))
		if err != nil {
			t.Fatal(err)
		}
	}
	checkFiles := func(n int) []string {
		files, err := filepath.Glob(path + "*")
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != n {
			t.Fatalf("unexpected number of files: %d", len(files))
		}
		return files
	}
	read := func(s string) string {
		b, err := ioutil.ReadFile(s)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	write("1234567")
	files := checkFiles(1)
	if files[0] != path {
		t.Fatal("unexpected rotation")
	}

	write("89")
	files = checkFiles(1)
	if files[0] != path+".0" {
		t.Fatal("unexpected rotation")
	}
	if s := read(files[0]); s != "123456789" {
		t.Fatalf("unexpected data: %s", s)
	}

	write("abcdefgh")
	write("ijklmnop")
	write("qrst")
	files = checkFiles(3)
	if !reflect.DeepEqual(files, []string{path, path + ".0", path + ".1"}) {
		t.Fatalf("unexpected files: %v", files)
	}
	for i, expected := range []string{"qrst", "ijklmnop", "abcdefgh"} {
		if s := read(files[i]); s != expected {
			t.Fatalf("unexpected data: %s", s)
		}
	}
	w.Close()

	// Open an existing file
	w = NewWriter(path, 8, 2)
	write("uvwx")
	write("yz")
	files = checkFiles(3)
	if !reflect.DeepEqual(files, []string{path, path + ".0", path + ".1"}) {
		t.Fatalf("unexpected files: %v", files)
	}
	for i, expected := range []string{"yz", "qrstuvwx", "ijklmnop"} {
		if s := read(files[i]); s != expected {
			t.Fatalf("unexpected data: %s", s)
		}
	}
}
