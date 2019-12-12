package gud

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path"
)

const ObjectsDirName string = "objects"

type Object struct {
	Name string
	Size int64
	Mode os.FileMode
	Type string
}

func InitObjectsDir(name string) error {
	return os.Mkdir(path.Join(name, ObjectsDirName), os.ModeDir)
}

func CrerteBlob(name string) ([]byte, error) {
	dst, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	h := sha1.New()
	_, err = fmt.Fprintf(h, name)
	if err != nil {
		return nil, err
	}
	zipWriter := zlib.NewWriter(io.MultiWriter(dst, h))

	src, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(zipWriter, src)
	if err != nil {
		return nil, err
	}

	err = src.Close()
	if err != nil {
		return nil, err
	}

	err = dst.Close()
	if err != nil {
		return nil, err
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
