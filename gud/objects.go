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

type objectType int
type objectHash [sha1.Size]byte

const (
	typeBlob    objectType = 0
	typeTree    objectType = 1
	typeVersion objectType = 2
)

type object struct {
	Name string
	Hash objectHash
	Type objectType
}

type tree []object

// Version is a representation of a project version.
type Version struct {
	Message string
	Time    time.Time
	Tree    objectHash
	prev    *objectHash
}

// HasPrev returns true if the version has a predecessor.
func (v Version) HasPrev() bool {
	return v.prev != nil
}

type dirStructure struct {
	Name    string
	Objects tree
	Dirs    []dirStructure
}

func initObjectsDir(rootPath string) (*objectHash, error) {
	err := os.Mkdir(filepath.Join(rootPath, objectsDirPath), os.ModeDir)
	if err != nil {
		return nil, err
	}

	tree, err := createTree(rootPath, "", tree{})
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = gob.NewEncoder(&buffer).Encode(Version{
		Tree:    tree.Hash,
		Message: initialCommitName,
		Time:    time.Now(),
		prev:    nil,
	})
	if err != nil {
		return nil, err
	}

	return createObject(rootPath, initialCommitName, &buffer)
}

func createBlob(rootPath, relPath string) (*objectHash, error) {
	src, err := os.Open(filepath.Join(rootPath, relPath))
	if err != nil {
		return nil, err
	}

	hash, err := createObject(rootPath, relPath, src)
	if err != nil {
		return nil, err
	}

	err = src.Close()
	if err != nil {
		return nil, err
	}

	return hash, err
}

func createTree(rootPath, relPath string, tree tree) (*object, error) {
	return createGobObject(rootPath, relPath, tree, typeTree)
}

func createVersion(rootPath, relPath string, version Version) (*object, error) {
	return createGobObject(rootPath, relPath, version, typeVersion)
}

func createGobObject(rootPath, relPath string, obj interface{}, objectType objectType) (*object, error) {
	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(obj)
	if err != nil {
		return nil, err
	}

	hash, err := createObject(rootPath, relPath, &buffer)
	if err != nil {
		return nil, err
	}
	return &object{
		Name: filepath.Base(relPath),
		Hash: *hash,
		Type: objectType,
	}, nil
}

func createObject(rootPath, relPath string, src io.Reader) (*objectHash, error) {
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
	var ret objectHash
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

func writeHead(rootPath string, hash objectHash) error {
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

func loadHead(rootPath string) (*objectHash, error) {
	head, err := os.Open(filepath.Join(rootPath, headFileName))
	if err != nil {
		return nil, err
	}

	var hash objectHash
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

func loadTree(rootPath string, hash objectHash, ret interface{}) error {
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

func addToStructure(structure *dirStructure, name string, hash objectHash) {
	dirs := strings.Split(name, string(os.PathSeparator))
	current := structure

	for _, dir := range dirs[:len(dirs)-1] {
		ind := sort.Search(len(current.Dirs), func(i int) bool {
			return dir <= current.Dirs[i].Name
		})
		if ind >= len(current.Dirs) || current.Dirs[ind].Name != dir {
			next := dirStructure{Name: dir}

			current.Dirs = append(current.Dirs, dirStructure{})
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

	current.Objects = append(current.Objects, object{})
	copy(current.Objects[ind+1:], current.Objects[ind:])
	current.Objects[ind] = object{
		Name: name,
		Hash: hash,
		Type: typeBlob,
	}
}

func buildTree(rootPath, relPath string, root dirStructure, prev tree) (*object, error) {
	newTree := make(tree, 0, len(prev)+len(root.Objects)+len(root.Dirs))
	copy(newTree, prev)

	for _, dir := range root.Dirs {
		var tree tree
		ind, found := searchTree(newTree, dir.Name)
		if found {
			err := loadTree(rootPath, newTree[ind].Hash, &tree)
			if err != nil {
				return nil, err
			}
		}

		obj, err := buildTree(rootPath, filepath.Join(relPath, dir.Name), dir, tree)
		if err != nil {
			return nil, err
		}
		newTree = append(newTree, object{})
		copy(newTree[ind+1:], newTree[ind:])
		newTree[ind] = *obj
	}

	for _, obj := range root.Objects {
		ind, found := searchTree(newTree, obj.Name)
		if !found { // file is not yet added
			newTree = append(newTree, object{})
			copy(newTree[ind+1:], newTree[ind:]) // keep the slice sorted
		}
		newTree[ind] = obj // update entry if the file was already added
	}

	return createTree(rootPath, relPath, newTree)
}

func searchTree(tree tree, name string) (int, bool) {
	l := len(tree)
	ind := sort.Search(l, func(i int) bool {
		return name <= tree[i].Name
	})

	return ind, ind < l && name == tree[ind].Name
}
