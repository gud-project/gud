package gud

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const DirectoryType string = "04000"

type Object struct {
	Name string
	Size int64
	Mode os.FileMode
	Type string
}

func (o Object) getHash() ([]byte, error) {
	h := sha1.New()
	fmt.Fprintf(h, "%s %d", o.Type, string(o.Size))

	f, err := os.Open(o.Name)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(h, f)
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

type ObjectDir struct {
	objects []Object
}

func InitObjectDir(path string) error {
	return nil
}

func (od ObjectDir) NewTree(Name string) error {
	info, err := os.Lstat(Name)
	if err != nil {
		return err
	}

	od.objects = append(od.objects, Object{Name, info.Size(), info.Mode(), "tree"})
	return nil
}

func (od ObjectDir) NewBlob(Name string) error {
	info, err := os.Lstat(Name)
	if err != nil {
		return err
	}

	od.objects = append(od.objects, Object{Name, info.Size(), info.Mode(), "blob"})

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	f, err := os.Open(Name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	w.Close()

	return nil
}
