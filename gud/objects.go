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
	"time"
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
	Tree    ObjectHash
	Message string
	Time    time.Time
	Prev    *ObjectHash
}

type DirStructure struct {
	Name    string
	Objects Tree
	Dirs    []DirStructure
}

func InitObjectsDir(rootPath string) (*ObjectHash, error) {
	err := os.Mkdir(filepath.Join(rootPath, objectsDirPath), os.ModeDir)
	if err != nil {
		return nil, err
	}

	tree, err := CreateTree(rootPath, "", Tree{})
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = gob.NewEncoder(&buffer).Encode(Version{
		Tree:    tree.Hash,
		Message: initialCommitName,
		Time:    time.Now(),
		Prev:    nil,
	})
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

func WriteHead(rootPath string, hash ObjectHash) error {
	head, err := os.Create(filepath.Join(rootPath, headFileName))
	if err != nil {
		return err
	}

	_, err = head.Write(hash[:])
	if err != nil {
		return err
	}

	return head.Close()
}

func LoadHead(rootPath string) (*ObjectHash, error) {
	head, err := os.Open(filepath.Join(rootPath, headFileName))
	if err != nil {
		return nil, err
	}

	var hash ObjectHash
	_, err = head.Read(hash[:])
	if err != nil {
		return nil, err
	}

	err = head.Close()
	if err != nil {
		return nil, err
	}

	return &hash, nil
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

	err = f.Close()
	if err != nil {
		return nil
	}

	return gob.NewDecoder(zip).Decode(ret)
}

func AddToStructure(dirStructure *DirStructure, name string, hash ObjectHash) {
	dirs := strings.Split(name, string(os.PathSeparator))
	current := dirStructure

	for _, dir := range dirs[:len(dirs)-1] {
		ind := sort.Search(len(current.Dirs), func(i int) bool {
			return dir <= current.Dirs[i].Name
		})
		if ind >= len(current.Dirs) || current.Dirs[ind].Name != dir {
			next := DirStructure{Name: dir}

			current.Dirs = append(current.Dirs, DirStructure{})
			copy(current.Dirs[ind+1:], current.Dirs[ind:])
			current.Dirs[ind] = next
		}
		current = &current.Dirs[ind]
	}

	// assume name is not in objects
	ind, found := searchTree(current.Objects, name)
	if found {
		panic("???")
	}

	current.Objects = append(current.Objects, Object{})
	copy(current.Objects[ind+1:], current.Objects[ind:])
	current.Objects[ind] = Object{
		Name: name,
		Hash: hash,
		Type: typeBlob,
	}
}

func BuildTree(rootPath, relPath string, root DirStructure, prev Tree) (*Object, error) {
	newTree := make(Tree, 0, len(prev)+len(root.Objects)+len(root.Dirs))
	copy(newTree, prev)

	for _, dir := range root.Dirs {
		var tree Tree
		ind, found := searchTree(newTree, dir.Name)
		if found {
			err := LoadTree(rootPath, newTree[ind].Hash, &tree)
			if err != nil {
				return nil, err
			}
		}

		obj, err := BuildTree(rootPath, filepath.Join(relPath, dir.Name), dir, tree)
		if err != nil {
			return nil, err
		}
		newTree = append(newTree, Object{})
		copy(newTree[ind+1:], newTree[ind:])
		newTree[ind] = *obj
	}

	for _, obj := range root.Objects {
		ind, found := searchTree(newTree, obj.Name)
		if !found { // file is not yet added
			newTree = append(newTree, Object{})
			copy(newTree[ind+1:], newTree[ind:]) // keep the slice sorted
		}
		newTree[ind] = obj // update entry if the file was already added
	}

	return CreateTree(rootPath, relPath, newTree)
}

func searchTree(tree Tree, name string) (int, bool) {
	l := len(tree)
	ind := sort.Search(l, func(i int) bool {
		return name <= tree[i].Name
	})

	return ind, ind < l && name == tree[ind].Name
}
