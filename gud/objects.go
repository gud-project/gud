package gud

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const objectsDirPath string = gudPath + "/objects"
const hashLen = 2 * sha1.Size
const initialCommitName string = "initial commit"

type ObjectType int
type ObjectHash [sha1.Size]byte

const (
	typeBlob ObjectType = 0
	typeTree ObjectType = 1
)

type Object struct {
	Name string
	Hash ObjectHash
	Type ObjectType
}

type Version struct {
	Tree    []Object
	Message string
}

func InitObjectsDir(rootPath string) (*ObjectHash, error) {
	err := os.Mkdir(filepath.Join(rootPath, objectsDirPath), os.ModeDir)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = gob.NewEncoder(&buffer).Encode(Version{[]Object{}, initialCommitName})
	if err != nil {
		return nil, err
	}

	return CreateObject(rootPath, initialCommitName, &buffer)
}

func CreateBlob(rootPath, relPath string) (*ObjectHash, error) {
	src, err := os.Open(filepath.Join(rootPath, relPath))
	if err != nil {
		return nil, err
	}

	hash, err := CreateObject(rootPath, relPath, src)
	if err != nil {
		return nil, err
	}

	err = src.Close()
	if err != nil {
		return nil, err
	}

	return hash, err
}

func CreateObject(rootPath, relPath string, src io.Reader) (*ObjectHash, error) {
	var zipData bytes.Buffer

	hash := sha1.New()
	_, err := fmt.Fprintf(hash, relPath)
	if err != nil {
		return nil, err
	}

	// use compressed data for both the object content and the hash
	zipWriter := zlib.NewWriter(io.MultiWriter(&zipData, hash))

	_, err = io.Copy(zipWriter, src)
	if err != nil {
		return nil, err
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	sum := hash.Sum(nil)
	var ret ObjectHash
	copy(ret[:], sum)

	// Create the blob file
	var objName [hashLen]byte
	hex.Encode(objName[:], sum) // Get the hash of the file

	dst, err := os.Create(filepath.Join(rootPath, objectsDirPath, string(objName[:])))
	if err != nil {
		return nil, err
	}

	_, err = zipData.WriteTo(dst)
	if err != nil {
		return nil, err
	}

	err = dst.Close()
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
