package gud

import (
	"container/list"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const branchesDirPath = gudPath + "/branches"
const headFileName = gudPath + "/head"
const firstBranchName = "master"

type Head struct {
	IsDetached bool
	Branch     string
	Hash       ObjectHash
	MergedHash *ObjectHash
}

func (p Project) CreateBranch(name string) error {
	if name == "" {
		return Error{"branch name cannot be empty"}
	}

	if strings.ContainsRune(name, os.PathSeparator) {
		err := os.MkdirAll(filepath.Join(p.Path, branchesDirPath, filepath.Dir(name)), os.ModeDir)
		if err != nil {
			return err
		}
	}

	head, err := loadHead(p.Path)
	if err != nil {
		return err
	}

	hash, err := getCurrentHash(p.Path, *head)
	if err != nil {
		return err
	}

	return dumpBranch(p.Path, name, *hash)
}

func (p Project) CheckoutBranch(branch string) error {
	hash, err := loadBranch(p.Path, branch)
	if err != nil {
		return err
	}

	return p.Checkout(*hash)
}

func (p Project) Checkout(hash ObjectHash) error {
	err := p.assertNoChanges()
	if err != nil {
		return err
	}

	version, err := loadVersion(p.Path, hash)
	if err != nil {
		return err
	}
	tree, err := loadTree(p.Path, version.TreeHash)
	if err != nil {
		return err
	}
	err = removeChanges(p.Path, tree)
	if err != nil {
		return err
	}

	head, err := loadHead(p.Path)
	if err != nil {
		return err
	}

	return dumpHead(p.Path, Head{
		IsDetached: true,
		Hash:       hash,
		Branch:     head.Branch,
	})
}

func (p Project) MergeBranch(from string) (*Version, error) {
	hash, err := loadBranch(p.Path, from)
	if err != nil {
		return nil, err
	}

	return p.merge(*hash, from)
}

func (p Project) MergeHash(from ObjectHash) (*Version, error) {
	version, err := loadVersion(p.Path, from)
	if err != nil {
		return nil, err
	}

	return p.merge(from, fmt.Sprintf("\"%s\"", version.Message))
}

func (p Project) ListBranches(fn func(branch string) error) error {
	return filepath.Walk(filepath.Join(p.Path, branchesDirPath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			err = fn(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (p Project) merge(from ObjectHash, name string) (*Version, error) {
	err := p.assertNoChanges()
	if err != nil {
		return nil, err
	}

	head, err := loadHead(p.Path)
	if err != nil {
		return nil, err
	}
	if head.IsDetached {
		return nil, Error{"cannot merge while head is detached"}
	}

	to, err := getCurrentHash(p.Path, *head)
	if err != nil {
		return nil, err
	}

	toVersion, err := loadVersion(p.Path, *to)
	if err != nil {
		return nil, err
	}
	fromVersion, err := loadVersion(p.Path, from)
	if err != nil {
		return nil, err
	}

	oldToNew, err := isDescendent(p.Path, *to, from)
	if err != nil {
		return nil, err
	}
	if oldToNew {
		return toVersion, nil
	}

	newToOld, err := isDescendent(p.Path, from, *to)
	if err != nil {
		return nil, err
	}
	if newToOld {
		err = dumpBranch(p.Path, head.Branch, from)
		if err != nil {
			return nil, err
		}
		return fromVersion, nil
	}

	base := from
	foundBase := false
	for !foundBase {
		baseVersion, err := loadVersion(p.Path, base)
		if err != nil {
			return nil, err
		}
		if !baseVersion.HasPrev() {
			panic("???")
		}

		base = *baseVersion.prev
		foundBase, err = isDescendent(p.Path, *to, base)
		if err != nil {
			return nil, err
		}
	}

	baseVersion, err := loadVersion(p.Path, base)
	if err != nil {
		return nil, err
	}
	toTree, err := loadTree(p.Path, toVersion.TreeHash)
	if err != nil {
		return nil, err
	}
	fromTree, err := loadTree(p.Path, fromVersion.TreeHash)
	if err != nil {
		return nil, err
	}
	baseTree, err := loadTree(p.Path, baseVersion.TreeHash)
	if err != nil {
		return nil, err
	}

	tree, conflicts, err := mergeTrees(p.Path, ".", toTree, fromTree, head.Branch, name, baseTree)
	if err != nil {
		return nil, err
	}

	if conflicts != nil {
		index := make([]indexEntry, 0, conflicts.Len())
		for p := conflicts.Front(); p != nil; p = p.Next() {
			index = append(index, indexEntry{
				Name:  p.Value.(string),
				State: StateConflict,
			})
		}

		err = dumpHead(p.Path, Head{
			IsDetached: false,
			Branch:     head.Branch,
			MergedHash: &from,
		})
		if err != nil {
			return nil, err
		}
		return nil, Error{"there are merge conflicts. please solve them and save the changes"}
	}

	treeObj, err := createTree(p.Path, ".", tree)
	if err != nil {
		return nil, err
	}

	return saveVersion(
		p.Path, fmt.Sprintf("merged %s into %s", name, head.Branch),
		head.Branch, treeObj.Hash, to, &from)
}

func removeChanges(rootPath string, tree tree) error {
	return compareTree(
		rootPath, ".", tree, []indexEntry{},
		func(relPath string, state FileState, hash *ObjectHash, isDir bool) error {
			path := filepath.Join(rootPath, relPath)
			if isDir && state == StateNew {
				return os.Remove(path)
			}
			if isDir {
				return os.Mkdir(path, os.ModeDir)
			}
			if state == StateNew {
				return os.Remove(path)
			}
			return extractBlob(rootPath, relPath, *hash)
		},
	)
}

func getCurrentHash(rootPath string, head Head) (*ObjectHash, error) {
	if head.IsDetached {
		return &head.Hash, nil
	}
	return loadBranch(rootPath, head.Branch)
}

func isDescendent(rootPath string, new, old ObjectHash) (bool, error) {
	for new != old {
		version, err := loadVersion(rootPath, new)
		if err != nil {
			return false, err
		}
		if !version.HasPrev() {
			return false, nil
		}
		new = *version.prev
	}

	return true, nil
}

func (p Project) assertNoChanges() error {
	return p.Status(
		func(relPath string, state FileState) error {
			return Error{"the index must be empty when checking out"}
		},
		func(relPath string, state FileState) error {
			return Error{"uncommitted changes must be cleaned before checking out"}
		},
	)
}

func mergeTrees(rootPath, relPath string, to, from tree, toName, fromName string, base tree) (tree, *list.List, error) {
	res := make(tree, len(to), len(to)+len(from))
	conflicts := list.New()
	copy(res, to)

	toInd := 0
	fromInd := 0
	for toInd < len(to) && fromInd < len(from) {
		toObj := to[toInd]
		fromObj := from[fromInd]

		if toObj.Name < fromObj.Name { // new object in target
			toInd++
		} else if fromObj.Name < toObj.Name { // new object in merged
			res = append(res, object{})
			copy(res[fromInd+1:], res[fromInd:])
			res[fromInd] = fromObj

			fromInd++
		} else {
			if toObj.Hash != fromObj.Hash {
				baseInd, found := searchTree(base, toObj.Name)
				if found {
					baseObj := base[baseInd]
					if toObj.Hash != baseObj.Hash && fromObj.Hash != baseObj.Hash { // conflicting changes
						mergedObj, newConflicts, err := mergeDiff(rootPath, relPath, toObj, fromObj, toName, fromName, &baseObj)
						if err != nil {
							return nil, nil, err
						}
						if newConflicts != nil {
							conflicts.PushFrontList(newConflicts)
						} else {
							res[toInd] = *mergedObj
						}

					} else if fromObj.Hash != baseObj.Hash { // change in merged
						res[toInd] = fromObj
					}
				} else { // conflicting changes
					mergedObj, newConflicts, err := mergeDiff(rootPath, relPath, toObj, fromObj, toName, fromName, nil)
					if err != nil {
						return nil, nil, err
					}
					if newConflicts != nil {
						conflicts.PushFrontList(newConflicts)
					} else {
						res[toInd] = *mergedObj
					}
				}
			}

			toInd++
			fromInd++
		}
	}
	for ; fromInd < len(from); fromInd++ {
		res = append(res, from[fromInd])
	}

	if conflicts.Len() > 0 {
		return nil, conflicts, nil
	}
	return res, nil, nil
}

func mergeDiff(
	rootPath, parentPath string,
	to, from object,
	toName, fromName string,
	base *object) (*object, *list.List, error) {
	relPath := filepath.Join(parentPath, to.Name)

	if to.Type != from.Type {
		return nil, nil, Error{"cannot merge directory and file: " + relPath}
	}
	if to.Type == typeBlob {
		err := writeConflict(rootPath, relPath, to.Hash, from.Hash, toName, fromName)
		if err != nil {
			return nil, nil, err
		}

		conflicts := list.New()
		conflicts.PushFront(relPath)
		return nil, conflicts, nil
	}

	toTree, err := loadTree(rootPath, to.Hash)
	if err != nil {
		return nil, nil, err
	}
	fromTree, err := loadTree(rootPath, from.Hash)
	if err != nil {
		return nil, nil, err
	}
	var baseTree tree
	if base != nil && base.Type == typeTree {
		baseTree, err = loadTree(rootPath, base.Hash)
	}

	newTree, conflicts, err := mergeTrees(rootPath, relPath, toTree, fromTree, toName, fromName, baseTree)
	if err != nil {
		return nil, nil, err
	}
	if conflicts != nil {
		return nil, conflicts, nil
	}

	newObj, err := createTree(rootPath, relPath, newTree)
	if err != nil {
		return nil, nil, err
	}

	return newObj, nil, nil
}

func writeConflict(
	rootPath, relPath string,
	to, from ObjectHash,
	toName, fromName string) (err error) {
	toText, err := readBlob(rootPath, to)
	if err != nil {
		return
	}
	fromText, err := readBlob(rootPath, from)
	if err != nil {
		return
	}

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

	dmp := diffmatchpatch.New()
	wSrc, wDst, wArr := dmp.DiffLinesToChars(toText, fromText)
	for _, diff := range dmp.DiffCharsToLines(dmp.DiffMain(wSrc, wDst, false), wArr) {
		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			_, err = dst.WriteString(diff.Text)
		case diffmatchpatch.DiffDelete:
			_, err = writeChange(dst, "Old", toName, diff.Text)
		case diffmatchpatch.DiffInsert:
			_, err = writeChange(dst, "New", fromName, diff.Text)
		}
		if err != nil {
			return
		}
	}

	return
}

func writeChange(w io.Writer, changeType, name, change string) (int, error) {
	label := fmt.Sprintf("{{{ %s change from %s {{{", changeType, name)
	return fmt.Fprintf(w, "%s\n%s\n%s", label, change, strings.Repeat("}", len(label)))
}

func initBranches(rootPath string, firstHash ObjectHash) error {
	err := os.Mkdir(filepath.Join(rootPath, branchesDirPath), os.ModeDir)
	if err != nil {
		return err
	}

	err = dumpBranch(rootPath, firstBranchName, firstHash)
	if err != nil {
		return err
	}

	return dumpHead(rootPath, Head{IsDetached: false, Branch: firstBranchName})
}

func dumpBranch(rootPath string, name string, hash ObjectHash) (err error) {
	file, err := os.Create(filepath.Join(rootPath, branchesDirPath, name))
	if err != nil {
		return
	}
	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = file.Write(hash[:])
	return
}

func loadBranch(rootPath, name string) (*ObjectHash, error) {
	var hash ObjectHash

	file, err := os.Open(filepath.Join(rootPath, branchesDirPath, name))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	n, err := file.Read(hash[:])
	if err != nil {
		return nil, err
	}
	if n != len(hash) {
		return nil, Error{"branch is corrupted"}
	}

	return &hash, nil
}

func dumpHead(rootPath string, head Head) (err error) {
	file, err := os.Create(filepath.Join(rootPath, headFileName))
	if err != nil {
		return err
	}
	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	return gob.NewEncoder(file).Encode(head)
}

func loadHead(rootPath string) (*Head, error) {
	file, err := os.Open(filepath.Join(rootPath, headFileName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var head Head
	err = gob.NewDecoder(file).Decode(&head)
	if err != nil {
		return nil, err
	}

	return &head, nil
}
