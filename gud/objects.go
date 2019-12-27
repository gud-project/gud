package gud

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
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

var nullHash objectHash

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

	obj, err := createVersion(rootPath, Version{
		Message: initialCommitName,
		Time:    time.Now(),
		Tree:    tree.Hash,
		prev:    nil,
	})
	if err != nil {
		return nil, err
	}

	return &obj.Hash, err
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

func createVersion(rootPath string, version Version) (*object, error) {
	return createGobObject(rootPath, version.Message, version, typeVersion)
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
	return ioutil.WriteFile(filepath.Join(rootPath, headFileName), hash[:], 0644)
}

func loadHead(rootPath string) (*objectHash, *Version, error) {
	head, err := os.Open(filepath.Join(rootPath, headFileName))
	if err != nil {
		return nil, nil, err
	}
	defer head.Close()

	var hash objectHash
	_, err = head.Read(hash[:])
	if err != nil {
		return nil, nil, err
	}

	var version Version
	err = loadTree(rootPath, hash, &version)
	if err != nil {
		return nil, nil, err
	}

	return &hash, &version, nil
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

func findObject(rootPath, relPath string) (*objectHash, error) {
	dirs := strings.Split(relPath, string(os.PathSeparator))
	_, version, err := loadHead(rootPath)
	if err != nil {
		return nil, err
	}

	hash := version.Tree
	for _, name := range dirs {
		var tree tree
		err := loadTree(rootPath, hash, &tree)
		if err != nil {
			return nil, err
		}

		ind, found := searchTree(tree, name)
		if !found {
			return nil, nil
		}
		hash = tree[ind].Hash
	}

	return &hash, nil
}

func compareToObject(rootPath, relPath string, hash objectHash) (bool, error) {
	const bufSiz = 1024

	file, err := os.Open(filepath.Join(rootPath, relPath))
	if err != nil {
		return false, err
	}
	defer file.Close()

	obj, err := os.Open(filepath.Join(rootPath, objectsDirPath, hex.EncodeToString(hash[:])))
	if err != nil {
		return false, err
	}
	defer obj.Close()

	unzip, err := zlib.NewReader(obj)
	if err != nil {
		return false, err
	}
	defer unzip.Close()

	var buf1, buf2 [bufSiz]byte
	for {
		n1, err1 := file.Read(buf1[:])
		if err1 != nil && err1 != io.EOF {
			return false, err1
		}
		n2, err2 := unzip.Read(buf2[:])
		if err2 != nil && err2 != io.EOF {
			return false, err2
		}

		if err1 == io.EOF || err2 == io.EOF {
			n1, err1 = file.Read(buf1[:])
			n2, err2 = unzip.Read(buf2[:])
			return n1 == 0 && n2 == 0 && err1 == io.EOF && err2 == io.EOF, nil
		}
		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}
	}
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

		if obj != nil {
			newTree = append(newTree, object{})
			copy(newTree[ind+1:], newTree[ind:])
			newTree[ind] = *obj
		}
	}

	for _, obj := range root.Objects {
		ind, found := searchTree(newTree, obj.Name)
		if obj.Hash == nullHash { // file is to be removed
			if !found {
				panic("???")
			}
			copy(newTree[ind:], newTree[ind+1:])
			newTree = newTree[:len(newTree)-1]

		} else {
			if !found { // file is not yet added
				newTree = append(newTree, object{})
				copy(newTree[ind+1:], newTree[ind:]) // keep the slice sorted
			}

			newTree[ind] = obj // update entry if the file was already added
		}
	}

	if len(newTree) == 0 {
		return nil, nil
	}
	return createTree(rootPath, relPath, newTree)
}

func walkBlobs(rootPath, relPath string, root tree, fn func(relPath string) error) error {
	for _, obj := range root {
		objRelPath := filepath.Join(relPath, obj.Name)
		if obj.Type == typeTree {
			var inner tree
			err := loadTree(rootPath, obj.Hash, &inner)
			if err != nil {
				return err
			}
			err = walkBlobs(rootPath, objRelPath, inner, fn)
			if err != nil {
				return err
			}
		} else {
			err := fn(objRelPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func searchTree(tree tree, name string) (int, bool) {
	l := len(tree)
	ind := sort.Search(l, func(i int) bool {
		return name <= tree[i].Name
	})

	return ind, ind < l && name == tree[ind].Name
}
