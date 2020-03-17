package gud

import (
	"compress/zlib"
	"container/list"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const blobContentType = "application/x-gud-blob"
const treeContentType = "application/x-gud-tree"
const versionContentType = "application/x-gud-version"

type InputError Error
func (e InputError) Error() string {
	return e.s
}

func (p Project) PushBranch(out io.Writer, branch string, start *ObjectHash) (boundary string, err error) {
	hash, err := p.GetBranch(branch)
	if err != nil {
		return
	}
	if hash == nil {
		return "", InputError{"branch does not exist"}
	}

	versions := list.New()
	err = getVersions(p.gudPath, *hash, start, versions)
	if err != nil {
		return "", err
	}

	writer := multipart.NewWriter(out)
	defer func() {
		cerr := writer.Close()
		if err == nil {
			err = cerr
		}
	}()

	for e := versions.Back(); e != nil; e = e.Prev() {
		hash := e.Value.(ObjectHash)
		err = pushVersion(p.gudPath, writer, hash)
		if err != nil {
			return "", err
		}
	}

	return writer.Boundary(), nil
}

func getVersions(gudPath string, hash ObjectHash, start *ObjectHash, nexts *list.List) error {
	if start != nil && hash == *start {
		return nil
	}
	for e := nexts.Front(); e != nil; e = e.Next() {
		if e.Value.(ObjectHash) == hash {
			return nil
		}
	}

	nexts.PushBack(hash)

	version, err := loadVersion(gudPath, hash)
	if err != nil {
		return err
	}

	if version.HasPrev() {
		err = getVersions(gudPath, *version.prev, start, nexts)
		if err != nil {
			return err
		}

		if version.IsMergeVersion() {
			err = getVersions(gudPath, *version.merged, start, nexts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func pushVersion(gudPath string, writer *multipart.Writer, hash ObjectHash) error {
	part, err := createPart(writer, hash, versionContentType)
	if err != nil {
		return err
	}

	src, err := os.Open(objectPath(gudPath, hash))
	if err != nil {
		return err
	}
	defer src.Close()

	version, err := versionFromReader(io.TeeReader(src, part))
	if err != nil {
		return err
	}

	return pushTree(gudPath, writer, version.TreeHash)
}

func pushTree(gudPath string, writer *multipart.Writer, hash ObjectHash) error {
	part, err := createPart(writer, hash, treeContentType)
	if err != nil {
		return err
	}

	src, err := os.Open(objectPath(gudPath, hash))
	if err != nil {
		return err
	}
	defer src.Close()

	var t tree
	err = readGobObject(io.TeeReader(src, part), &t)
	if err != nil {
		return err
	}

	for _, obj := range t {
		if obj.Type == typeTree {
			err = pushTree(gudPath, writer, obj.Hash)
		} else {
			err = pushBlob(gudPath, writer, obj.Hash)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func pushBlob(gudPath string, writer *multipart.Writer, hash ObjectHash) error {
	part, err := createPart(writer, hash, blobContentType)
	if err != nil {
		return err
	}

	src, err := os.Open(objectPath(gudPath, hash))
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = io.Copy(part, src)
	return err
}

func createPart(writer *multipart.Writer, hash ObjectHash, contentType string) (io.Writer, error) {
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, hash))
	header.Set("Content-Type", contentType)

	return writer.CreatePart(header)
}

func (p Project) PullBranch(branch string, in io.Reader, contentType string) error {
	return p.PullBranchFrom(branch, in, contentType, "")
}

func (p Project) PullBranchFrom(branch string, in io.Reader, contentType, user string) error {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return InputError{fmt.Sprintf("invalid content type: %s", contentType)}
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		return InputError{fmt.Sprintf("invalid content type: %s", contentType)}
	}

	currentHash, err := p.GetBranch(branch)
	if err != nil {
		return err
	}

	temp, err := createTempProject(p)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(temp.Path)
	}()

	files := list.New()
	objs := multipart.NewReader(in, params["boundary"])
	for {
		hash, err := pullVersion(temp.gudPath, user, objs, currentHash, files)
		if err != nil {
			return err
		}
		if hash == nil {
			break
		}
		currentHash = hash
	}

	for e := files.Front(); e != nil; e = e.Next() {
		name := e.Value.(string)
		err = copyFile(filepath.Join(temp.gudPath, objectsPath, name), filepath.Join(p.gudPath, objectsPath, name))
		if err != nil {
			return err
		}
	}

	if currentHash != nil {
		err = dumpBranch(p.gudPath, branch, *currentHash)
		if err != nil {
			return err
		}
	}

	return nil
}

func pullVersion(gudPath, user string, reader *multipart.Reader, prevHash *ObjectHash, files *list.List,
	) (hash *ObjectHash, err error) {
	part, err := reader.NextPart()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, InputError{"invalid multipart data"}
	}
	defer part.Close()

	hash, err = validatePart(gudPath, part, versionContentType)
	if err != nil {
		return
	}

	name := part.FileName()
	dst, err := os.Create(objectPath(gudPath, *hash))
	if err != nil {
		return
	}
	defer func() {
		cerr := dst.Close()
		if err == nil {
			err = cerr
		}
	}()

	current, err := versionFromReader(io.TeeReader(part, dst))
	if err != nil {
		return
	}

	prev, err := validateVersion(gudPath, user, *current, *hash, prevHash)
	if err != nil {
		return
	}

	files.PushBack(name)

	var prevTree tree
	if prev != nil {
		prevTree, err = loadTree(gudPath, prev.TreeHash)
		if err != nil {
			return
		}
	}

	err = pullTree(gudPath, reader, current.TreeHash, prevTree, files)
	return
}

func pullTree(
	gudPath string, reader *multipart.Reader, expectedHash ObjectHash, prev tree, files *list.List) error {
	part, err := reader.NextPart()
	if err != nil {
		return InputError{"invalid multipart data"}
	}
	defer part.Close()

	hash, err := validatePart(gudPath, part, treeContentType)
	if err != nil {
		return err
	}
	if *hash != expectedHash {
		return InputError{fmt.Sprintf("unexpected tree: expected %s, got %s", expectedHash, hash)}
	}

	name := part.FileName()
	dst, err := os.Create(objectPath(gudPath, *hash))
	if err != nil {
		return err
	}
	defer func() {
		cerr := dst.Close()
		if err == nil {
			err = cerr
		}
	}()

	var current tree
	err = readGobObject(io.TeeReader(part, dst), &current)
	if err != nil {
		return InputError{fmt.Sprintf("invalid tree object: %s", hash)}
	}

	if !sort.IsSorted(current) {
		return InputError{fmt.Sprintf("invalid tree: %s", hash)}
	}

	files.PushBack(name)

	for _, obj := range current {
		var prevObj *object
		for _, p := range prev {
			if p.Hash == obj.Hash {
				prevObj = &p
				break
			}
		}
		if prevObj != nil {
			if obj.Type != prevObj.Type {
				return InputError{fmt.Sprintf("invalid tree: %s", hash)}
			}
		} else {
			switch obj.Type {
			case typeBlob:
				err = pullBlob(gudPath, reader, obj.Hash, files)
				if err != nil {
					return err
				}

			case typeTree:
				var prevChild tree
				if ind, found := searchTree(prev, obj.Name); found {
					prevChild, err = loadTree(gudPath, prev[ind].Hash)
					if err != nil {
						return err
					}
				}

				err = pullTree(gudPath, reader, obj.Hash, prevChild, files)
				if err != nil {
					return err
				}

			default:
				return InputError{fmt.Sprintf("invalid tree: %s", hash)}
			}
		}
	}

	return nil
}

func pullBlob(gudPath string, reader *multipart.Reader, expectedHash ObjectHash, files *list.List) error {
	part, err := reader.NextPart()
	if err != nil {
		return InputError{"invalid multipart data"}
	}
	defer part.Close()

	hash, err := validatePart(gudPath, part, blobContentType)
	if err != nil {
		return err
	}
	if *hash != expectedHash {
		return InputError{fmt.Sprintf("unexpected blob: expected %s, got %s", expectedHash, hash)}
	}

	name := part.FileName()
	dst, err := os.Create(objectPath(gudPath, *hash))
	if err != nil {
		return err
	}
	defer func() {
		cerr := dst.Close()
		if err == nil {
			err = cerr
		}
	}()

	zip, err := zlib.NewReader(io.TeeReader(part, dst))
	if err != nil {
		return err
	}
	defer zip.Close()

	_, err = ioutil.ReadAll(zip)
	if err != nil {
		return err
	}

	files.PushBack(name)
	return nil
}

func validatePart(gudPath string, part *multipart.Part, expectedType string) (*ObjectHash, error) {
	name := part.FileName()
	var hash ObjectHash
	n, err := hex.Decode(hash[:], []byte(name))
	if err != nil || n != len(hash) {
		return nil, InputError{fmt.Sprintf("invalid file name: %s", name)}
	}

	contentType := part.Header.Get("Content-Type")
	if contentType != expectedType {
		return nil, InputError{
			fmt.Sprintf("invalid content type: expected %s, got %s", expectedType, contentType)}
	}

	_, err = os.Stat(objectPath(gudPath, hash))
	if !os.IsNotExist(err) {
		return nil, InputError{fmt.Sprintf("object already exists: %s", name)}
	}

	return &hash, nil
}

func validateVersion(rootPath, user string, v Version, hash ObjectHash, prevHash *ObjectHash) (*Version, error) {
	if user != "" && v.Author != user {
		return nil, InputError{fmt.Sprintf("expected user %s, got %s", user, v.Author)}
	}

	if prevHash == nil {
		if v.HasPrev() {
			return nil, InputError{fmt.Sprintf("expected first version: %s", hash)}
		}
		if v.IsMergeVersion() {
			return nil, InputError{fmt.Sprintf("invalid merge version: %s", hash)}
		}

		return nil, nil
	}

	prev, err := loadVersion(rootPath, *prevHash)
	if err != nil {
		return nil, err
	}

	if !v.HasPrev() {
		return nil, InputError{fmt.Sprintf("unexpected first version: %s", hash)}
	}
	if *v.prev != *prevHash {
		return nil, InputError{fmt.Sprintf("unexpected version: %s", hash)}
	}
	if !prev.Time.Before(v.Time) {
		return nil, InputError{fmt.Sprintf("invalid version time: %s", hash)}
	}
	if v.IsMergeVersion() {
		merged, err := loadVersion(rootPath, *v.merged)
		if os.IsNotExist(err) {
			return nil, InputError{fmt.Sprintf("invalid merge version: %s", hash)}
		}
		if err != nil {
			return nil, err
		}
		if !merged.Time.Before(v.Time) {
			return nil, InputError{fmt.Sprintf("invalid version time: %s", hash)}
		}
	}

	return prev, nil
}

func versionFromReader(in io.Reader) (*Version, error) {
	var v gobVersion
	err := readGobObject(in, &v)
	if err != nil {
		return nil, InputError{"invalid version object"}
	}

	ret := gobToVersion(v)
	return &ret, nil
}

func createTempProject(p Project) (temp *Project, err error) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tempDir)
		}
	}()

	dstGud := filepath.Join(tempDir, DefaultPath)
	dst := filepath.Join(dstGud, objectsPath)
	err = os.MkdirAll(dst, 0700)
	if err != nil {
		return
	}

	src := filepath.Join(p.gudPath, objectsPath)
	objs, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}
	for _, obj := range objs {
		err = copyFile(filepath.Join(src, obj.Name()), filepath.Join(dst, obj.Name()))
		if err != nil {
			return
		}
	}

	return &Project{tempDir, dstGud}, nil
}

func copyFile(srcPath, dstPath string) (err error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
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
