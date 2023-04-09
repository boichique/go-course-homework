package main

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"
)

func Test_printHashes(t *testing.T) {
	fakeFS := fstest.MapFS{
		"dir/file1": &fstest.MapFile{Data: []byte("hello in dir")},
		"dir/file2": &fstest.MapFile{Data: []byte("world in dir")},
		"file1":     &fstest.MapFile{Data: []byte("hello")},
		"file2":     &fstest.MapFile{Data: []byte("world")},
	}

	want := `
sha256:d698a2d966fe4bee7bcf0000c96b3fd938103cb32041da42512fdb2d67e6d3e9 dir/file1
sha256:cf6b5692f2ad668e0d0e4015d0fee9d4134d0ce44ce04759547bad02a61a34f0 dir/file2
sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824 file1
sha256:486ea46224d1bb4fb680f34f7c9ad96a8f24ec88be73ea8e5a6c65260e9cb8a7 file2
`

	var buf bytes.Buffer
	if err := printHashes(fakeFS, &buf, "."); err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(buf.String()) != strings.TrimSpace(want) {
		t.Errorf("printHashes() = %v, want %v", buf.String(), want)
	}
}
