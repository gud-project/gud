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
	"sort"
	"strings"
)

const objectsDirPath string = gudPath + "/objects"
const hashLen = 2 * sha1.Size
const initialCommitName string = "initial commit"

type ObjectType int
type ObjectHash [sha1.Size]byte

const (
	typeBlob    ObjectType = 0
	typeTree    ObjectType = 1
	typeVersion ObjectType = 2
)

type Object struct {
	Name string
	Hash ObjectHash
	Type ObjectType
}

type Tree []Object

type Version struct {
	Tree    Tree
	Message string
}

func InitObjectsDir(rootPath string) (*ObjectHash, error) {
	err := os.Mkdir(filepath.Join(rootPath, objectsDirPath), os.ModeDir)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = gob.NewEncoder(&buffer).Encode(Version{Tree{}, initialCommitName})
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

func CreateTree(rootPath, relPath string, tree Tree) (*Object, error) {
	return CreateGobObject(rootPath, relPath, tree, typeTree)
}

func CreateVersion(rootPath, relPath string, version Version) (*Object, error) {
	return CreateGobObject(rootPath, relPath, version, typeVersion)
}

func CreateGobObject(rootPath, relPath string, obj interface{}, objectType ObjectType) (*Object, error) {
	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(obj)
	if err != nil {
		return nil, err
	}

	hash, err := CreateObject(rootPath, relPath, &buffer)
	if err != nil {
		return nil, err
	}
	return &Object{
		Name: filepath.Base(relPath),
		Hash: *hash,
		Type: objectType,
	}, nil
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

func LoadTree(rootPath string, hash ObjectHash, ret interface{}) error {
	f, err := os.Open(filepath.Join(rootPath, objectsDirPath, hex.EncodeToString(hash[:])))
	if err != nil {
		return err
	}

	zip, err := zlib.NewReader(f)
	if err != nil {
		return err
	}

	return gob.NewDecoder(zip).Decode(ret)
}

func FindObjectParent(rootPath, relPath string, root []Object) (*[]Object, error) {
	// WIP
	for _, dir := range strings.Split(relPath, string(os.PathSeparator)) {
		ind := sort.Search(len(root), func(i int) bool {
			return root[i].Name == dir
		})
		if ind >= len(root) || root[ind].Name != dir {
			return nil, Error{"Object not found"}
		}

		err := LoadTree(rootPath, root[ind].Hash, &root)
		if err != nil {
			return nil, err
		}
	}

	return &root, nil
}
