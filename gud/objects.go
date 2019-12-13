package gud

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const objectsDirPath string = gudPath + "/objects"
const hashLen = 2 * sha1.Size

type Object struct {
	Name string
	Size int64
	Mode os.FileMode
	Type string
}

func InitObjectsDir(rootPath string) error {
	return os.Mkdir(filepath.Join(rootPath, objectsDirPath), os.ModeDir)
}

func CreateBlob(rootPath, relPath string) (*[hashLen]byte, error) {
	var dst bytes.Buffer

	hash := sha1.New()
	_, err := fmt.Fprintf(hash, relPath)
	if err != nil {
		return nil, err
	}

	// use compressed data for both the object content and the hash
	zipWriter := zlib.NewWriter(io.MultiWriter(&dst, hash))

	src, err := os.Open(filepath.Join(rootPath, relPath))
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

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	var ret [hashLen]byte
	hex.Encode(ret[:], hash.Sum(nil))

	obj, err := os.Create(filepath.Join(rootPath, objectsDirPath, string(ret[:])))
	if err != nil {
		return nil, err
	}

	_, err = dst.WriteTo(obj)
	if err != nil {
		return nil, err
	}

	err = obj.Close()
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
