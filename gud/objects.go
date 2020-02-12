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
const initialCommitName string = "initial commit"

type objectType int
type ObjectHash [sha1.Size]byte

func (h ObjectHash) String() string {
	return hex.EncodeToString(h[:])
}

var nullHash ObjectHash

const (
	typeBlob    objectType = 0
	typeTree    objectType = 1
	typeVersion objectType = 2
)

type object struct {
	Name string
	Hash ObjectHash
	Type objectType
}

type tree []object

func (t tree) Len() int {
	return len(t)
}

func (t tree) Less(i, j int) bool {
	return t[i].Name < t[j].Name
}

func (t tree) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Version is a representation of a project version.
type Version struct {
	Message  string
	Time     time.Time
	TreeHash ObjectHash
	prev     *ObjectHash
	merged   *ObjectHash
}

type gobVersion struct {
	Message  string
	Time     time.Time
	TreeHash ObjectHash
	Prev     *ObjectHash
	Merged   *ObjectHash
}

func init() {
	gob.Register(&indexFile{})
	gob.Register(&Head{})
	gob.Register(&gobVersion{})
	gob.Register(&tree{})
}

func versionToGob(v Version) gobVersion {
	return gobVersion{
		Message:  v.Message,
		Time:     v.Time,
		TreeHash: v.TreeHash,
		Prev:     v.prev,
		Merged:   v.merged,
	}
}

func gobToVersion(v gobVersion) Version {
	return Version{
		Message:  v.Message,
		Time:     v.Time,
		TreeHash: v.TreeHash,
		prev:     v.Prev,
		merged:   v.Merged,
	}
}


// HasPrev returns true if the version has a predecessor.
func (v Version) HasPrev() bool {
	return v.prev != nil
}

func (v Version) IsMergeVersion() bool {
	return v.merged != nil
}

type dirStructure struct {
	Name    string
	Objects tree
	Dirs    []dirStructure
}

func initObjectsDir(rootPath string) error {
	return os.Mkdir(filepath.Join(rootPath, objectsDirPath), dirPerm)
}

func objectPath(rootPath string, hash ObjectHash) string {
	return filepath.Join(rootPath, objectsDirPath, hash.String())
}

func saveVersion(rootPath, message, branch string, tree ObjectHash, prev, merged *ObjectHash) (*Version, error) {
	v := Version{
		Message:  message,
		Time:     time.Now(),
		TreeHash: tree,
		prev:     prev,
		merged:   merged,
	}

	obj, err := createVersion(rootPath, v)
	if err != nil {
		return nil, err
	}

	err = dumpBranch(rootPath, branch, obj.Hash)
	if err != nil {
		return nil, err
	}

	return &v, err
}

func createBlob(rootPath, relPath string) (*ObjectHash, error) {
	src, err := os.Open(filepath.Join(rootPath, relPath))
	if err != nil {
		return nil, err
	}
	defer src.Close()

	hash, err := createObject(rootPath, relPath, src)
	if err != nil {
		return nil, err
	}

	return hash, err
}

func createTree(rootPath, relPath string, tree tree) (*object, error) {
	return createGobObject(rootPath, relPath, tree, typeTree)
}

func createVersion(rootPath string, version Version) (*object, error) {
	return createGobObject(rootPath, version.Message, versionToGob(version), typeVersion)
}

func (v *Version) String() string {
	return fmt.Sprintf("Message: %s\nTime: %s\nHash: %s", v.Message, v.Time.Format("2006-01-02 15:04:05"), v.TreeHash.String())
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

func createObject(rootPath, relPath string, src io.Reader) (hash *ObjectHash, err error) {
	var zipData bytes.Buffer

	sha := sha1.New()
	_, err = fmt.Fprintf(sha, relPath)
	if err != nil {
		return
	}

	// use compressed data for both the object content and the hash
	zip := zlib.NewWriter(io.MultiWriter(&zipData, sha))
	defer func() {
		cerr := zip.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(zip, src)
	if err != nil {
		return
	}

	err = zip.Close()
	if err != nil {
		return
	}

	sum := sha.Sum(nil)
	var ret ObjectHash
	copy(ret[:], sum)

	// Create the blob file
	dst, err := os.Create(objectPath(rootPath, ret))
	if err != nil {
		return
	}
	defer func() {
		cerr := dst.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = zipData.WriteTo(dst)
	if err != nil {
		return
	}

	return &ret, nil
}

func readBlob(rootPath string, hash ObjectHash) (string, error) {
	src, err := os.Open(objectPath(rootPath, hash))
	if err != nil {
		return "", err
	}
	defer src.Close()

	zip, err := zlib.NewReader(src)
	if err != nil {
		return "", err
	}
	defer zip.Close()

	content, err := ioutil.ReadAll(zip)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func extractBlob(rootPath, relPath string, hash ObjectHash) (err error) {
	src, err := os.Open(objectPath(rootPath, hash))
	if err != nil {
		return
	}
	defer src.Close()

	zip, err := zlib.NewReader(src)
	if err != nil {
		return
	}
	defer zip.Close()

	dst, err := os.Create(filepath.Join(rootPath, relPath))
	if err != nil {
		return
	}
	defer func() {
		cerr := dst.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(dst, src)
	return
}

func readGobObject(in io.Reader, ret interface{}) error {
	zip, err := zlib.NewReader(in)
	if err != nil {
		return err
	}
	defer zip.Close()

	return gob.NewDecoder(zip).Decode(ret)
}

func loadGobObject(rootPath string, hash ObjectHash, ret interface{}) error {
	f, err := os.Open(objectPath(rootPath, hash))
	if err != nil {
		return err
	}
	defer f.Close()

	return readGobObject(f, ret)
}

func loadTree(rootPath string, hash ObjectHash) (tree, error) {
	var t tree

	err := loadGobObject(rootPath, hash, &t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func loadVersion(rootPath string, hash ObjectHash) (*Version, error) {
	var v gobVersion

	err := loadGobObject(rootPath, hash, &v)
	if err != nil {
		return nil, err
	}

	ret := gobToVersion(v)
	return &ret, nil
}

func findObject(rootPath, relPath string) (*ObjectHash, error) {
	dirs := strings.Split(relPath, string(os.PathSeparator))
	head, err := loadHead(rootPath)
	if err != nil {
		return nil, err
	}

	versionHash, err := getCurrentHash(rootPath, *head)
	if err != nil {
		return nil, err
	}

	version, err := loadVersion(rootPath, *versionHash)
	if err != nil {
		return nil, err
	}

	hash := version.TreeHash
	for _, name := range dirs {
		tree, err := loadTree(rootPath, hash)
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

func compareToObject(rootPath, relPath string, hash ObjectHash) (bool, error) {
	const bufSiz = 1024

	file, err := os.Open(filepath.Join(rootPath, relPath))
	if err != nil {
		return false, err
	}
	defer file.Close()

	obj, err := os.Open(objectPath(rootPath, hash))
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

func addToStructure(structure *dirStructure, relPath string, hash ObjectHash) {
	dirs := strings.Split(relPath, string(os.PathSeparator))
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

	name := dirs[len(dirs)-1]
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
	newTree := make(tree, len(prev), len(prev)+len(root.Objects)+len(root.Dirs))
	copy(newTree, prev)

	for _, dir := range root.Dirs {
		var tree tree
		ind, found := searchTree(newTree, dir.Name)
		if found {
			var err error
			tree, err = loadTree(rootPath, newTree[ind].Hash)
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

func walkBlobs(rootPath, relPath string, root tree, fn func(relPath string, obj object) error) error {
	for _, obj := range root {
		objRelPath := filepath.Join(relPath, obj.Name)

		if obj.Type == typeTree {
			inner, err := loadTree(rootPath, obj.Hash)
			if err != nil {
				return err
			}
			err = walkBlobs(rootPath, objRelPath, inner, fn)
			if err != nil {
				return err
			}
		}

		err := fn(objRelPath, obj)
		if err != nil {
			return err
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
