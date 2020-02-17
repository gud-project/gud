package gud

import (
	"bytes"
	"compress/zlib"
	"container/list"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const objectsPath string = "objects"
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

func initObjectsDir(gudPath string) error {
	return os.Mkdir(filepath.Join(gudPath, objectsPath), dirPerm)
}

func objectPath(gudPath string, hash ObjectHash) string {
	return filepath.Join(gudPath, objectsPath, hash.String())
}

func saveVersion(gudPath, message, branch string, tree ObjectHash, prev, merged *ObjectHash) (*Version, error) {
	v := Version{
		Message:  message,
		Time:     time.Now(),
		TreeHash: tree,
		prev:     prev,
		merged:   merged,
	}

	obj, err := createVersion(gudPath, v)
	if err != nil {
		return nil, err
	}

	err = dumpBranch(gudPath, branch, obj.Hash)
	if err != nil {
		return nil, err
	}

	return &v, err
}

func (p Project) createBlob(relPath string) (h *ObjectHash, err error) {
	src, err := os.Open(filepath.Join(p.Path, relPath))
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := newObjectWriter(relPath)
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
	if err != nil {
		return
	}

	return dst.Dump(p.gudPath)
}

func createTree(gudPath, relPath string, tree tree) (*object, error) {
	return createGobObject(gudPath, relPath, tree, typeTree)
}

func createVersion(gudPath string, version Version) (*object, error) {
	return createGobObject(gudPath, version.Message, versionToGob(version), typeVersion)
}

func createGobObject(gudPath, relPath string, ret interface{}, objectType objectType) (obj *object, err error) {
	w, err := newObjectWriter(relPath)
	if err != nil {
		return
	}
	defer func() {
		cerr := w.Close()
		if err == nil {
			err = cerr
		}
	}()

	err = gob.NewEncoder(w).Encode(ret)
	if err != nil {
		return
	}

	h, err := w.Dump(gudPath)
	if err != nil {
		return
	}

	return &object{
		Name: filepath.Base(relPath),
		Hash: *h,
		Type: objectType,
	}, nil
}

type objectWriter struct {
	io.WriteCloser
	data bytes.Buffer
	sha  hash.Hash
}

func newObjectWriter(name string) (*objectWriter, error) {
	w := &objectWriter{
		sha: sha1.New(),
	}
	_, err := fmt.Fprintf(w.sha, name)
	if err != nil {
		return nil, err
	}

	w.WriteCloser = zlib.NewWriter(io.MultiWriter(&w.data, w.sha))
	return w, nil
}

func (w *objectWriter) Dump(gudPath string) (h *ObjectHash, err error) {
	err = w.Close()
	if err != nil {
		return
	}

	sum := w.sha.Sum(nil)
	var ret ObjectHash
	copy(ret[:], sum)

	// Create the blob file
	dst, err := os.Create(objectPath(gudPath, ret))
	if err != nil {
		return
	}
	defer func() {
		cerr := dst.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = w.data.WriteTo(dst)
	if err != nil {
		return
	}

	return &ret, nil
}

func readBlob(gudPath string, hash ObjectHash) (string, error) {
	src, err := os.Open(objectPath(gudPath, hash))
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

func (p Project) extractBlob(relPath string, hash ObjectHash) (err error) {
	src, err := os.Open(objectPath(p.gudPath, hash))
	if err != nil {
		return
	}
	defer src.Close()

	zip, err := zlib.NewReader(src)
	if err != nil {
		return
	}
	defer zip.Close()

	dst, err := os.Create(filepath.Join(p.Path, relPath))
	if err != nil {
		return
	}
	defer func() {
		cerr := dst.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(dst, zip)
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

func loadGobObject(gudPath string, hash ObjectHash, ret interface{}) error {
	f, err := os.Open(objectPath(gudPath, hash))
	if err != nil {
		return err
	}
	defer f.Close()

	return readGobObject(f, ret)
}

func loadTree(gudPath string, hash ObjectHash) (tree, error) {
	var t tree

	err := loadGobObject(gudPath, hash, &t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func loadVersion(gudPath string, hash ObjectHash) (*Version, error) {
	var v gobVersion

	err := loadGobObject(gudPath, hash, &v)
	if err != nil {
		return nil, err
	}

	ret := gobToVersion(v)
	return &ret, nil
}

func (p Project) findObject(relPath string) (*object, error) {
	dirs := strings.Split(relPath, string(os.PathSeparator))
	versionHash, err := p.CurrentHash()
	if err != nil {
		return nil, err
	}

	version, err := loadVersion(p.gudPath, *versionHash)
	if err != nil {
		return nil, err
	}

	obj := object{".", version.TreeHash, typeTree}
	for _, name := range dirs {
		tree, err := loadTree(p.gudPath, obj.Hash)
		if err != nil {
			return nil, err
		}

		ind, found := searchTree(tree, name)
		if !found {
			return nil, nil
		}
		obj = tree[ind]
	}

	return &obj, nil
}

func (p Project) compareToObject(relPath string, hash ObjectHash) (bool, error) {
	const bufSiz = 1024

	file, err := os.Open(filepath.Join(p.Path, relPath))
	if err != nil {
		return false, err
	}
	defer file.Close()

	obj, err := os.Open(objectPath(p.gudPath, hash))
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
		// weird inconsistency: upon reaching end-of-file, the zip reader will return (n, io.EOF)
		// while the file reader will return (n, nil) and then (0, io.EOF) on the next read
		// shouldn't affect our code though
		n1, err1 := file.Read(buf1[:])
		if err1 != nil && err1 != io.EOF {
			return false, err1
		}
		n2, err2 := unzip.Read(buf2[:])
		if err2 != nil && err2 != io.EOF {
			return false, err2
		}

		if n1 != n2 {
			return false, nil
		}
		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}
		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
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

func buildTree(gudPath, relPath string, root dirStructure, prev tree) (*object, error) {
	newTree := make(tree, len(prev), len(prev)+len(root.Objects)+len(root.Dirs))
	copy(newTree, prev)

	for _, dir := range root.Dirs {
		var tree tree
		ind, found := searchTree(newTree, dir.Name)
		if found {
			var err error
			tree, err = loadTree(gudPath, newTree[ind].Hash)
			if err != nil {
				return nil, err
			}
		}

		obj, err := buildTree(gudPath, filepath.Join(relPath, dir.Name), dir, tree)
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
	return createTree(gudPath, relPath, newTree)
}

func removeVersion(gudPath string, last, afterLast Version, lastHash, afterLastHash ObjectHash) (err error) {
	afterLastObjs, err := listTree(gudPath, afterLast.TreeHash)
	if err != nil {
		return
	}

	lastObjs, err := listTree(gudPath, last.TreeHash)
	if err != nil {
		return
	}

	for lastE := lastObjs.Front(); lastE != nil; lastE = lastE.Next() {
		lastObj := lastE.Value.(ObjectHash)
		toRemove := true
		for afterLastE := afterLastObjs.Front(); afterLastE != nil; afterLastE = afterLastE.Next() {
			if afterLastE.Value.(ObjectHash) == lastObj {
				afterLastObjs.Remove(afterLastE)
				toRemove = false
				break
			}

			if toRemove {
				err = os.Remove(objectPath(gudPath, lastObj))
				if err != nil {
					return
				}
			}
		}
	}

	err = os.Remove(objectPath(gudPath, lastHash))
	if err != nil {
		return
	}

	dst, err := os.Create(objectPath(gudPath, afterLastHash))
	if err != nil {
		return
	}
	defer func() {
		cerr := dst.Close()
		if err == nil {
			err = cerr
		}
	}()

	afterLast.prev = nil
	return gob.NewEncoder(dst).Encode(versionToGob(afterLast))
}

func listTree(gudPath string, hash ObjectHash) (*list.List, error) {
	l := list.New()

	l.PushBack(hash)

	t, err := loadTree(gudPath, hash)
	if err != nil {
		return nil, err
	}

	for _, obj := range t {
		if obj.Type == typeTree {
			inner, err := listTree(gudPath, obj.Hash)
			if err != nil {
				return nil, err
			}
			l.PushBackList(inner)
		} else {
			l.PushBack(obj.Hash)
		}
	}

	return l, nil
}

func walkObjects(gudPath, relPath string, root tree, fn func(relPath string, obj object) error) error {
	for _, obj := range root {
		objRelPath := filepath.Join(relPath, obj.Name)

		if obj.Type == typeTree {
			inner, err := loadTree(gudPath, obj.Hash)
			if err != nil {
				return err
			}
			err = walkObjects(gudPath, objRelPath, inner, fn)
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
