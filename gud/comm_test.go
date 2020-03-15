package gud

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestProject_PushBranch(t *testing.T) {
	clientPath := filepath.Join(testDir, "client")
	serverPath := filepath.Join(testDir, "server")
	_ = os.Mkdir(clientPath, dirPerm)
	_ = os.Mkdir(serverPath, dirPerm)

	client, _ := Start(clientPath)
	server, _ := StartHeadless(serverPath)

	testData := "some test data"
	fileName := filepath.Join(clientPath, testFile)
	message := "add test file"

	_ = ioutil.WriteFile(fileName, []byte(testData), dirPerm)
	_ = client.Add(fileName)
	_, _ = client.Save(message)

	var buf bytes.Buffer
	boundary, err := client.PushBranch(&buf, FirstBranchName, nil)
	if err != nil {
		t.Fatal("failed to push branch: ", err)
	}

	err = server.PullBranch(FirstBranchName, &buf, "multipart/mixed; boundary="+boundary)
	if err != nil {
		t.Fatal("failed to pull branch:", err)
	}

	hash, err := loadBranch(server.gudPath, FirstBranchName)
	if err != nil {
		t.Fatal("failed to load branch:", err)
	}

	version, err := loadVersion(server.gudPath, *hash)
	if err != nil {
		t.Fatal("failed to load version:", err)
	}
	if version.Message != message {
		t.Fatal("invalid version")
	}

	tree, err := loadTree(server.gudPath, version.TreeHash)
	if err != nil {
		t.Fatal("failed to load tree:", err)
	}
	if len(tree) != 1 || tree[0].Name != testFile || tree[0].Type != typeBlob {
		t.Fatal("invalid tree")
	}

	data, err := readBlob(server.gudPath, tree[0].Hash)
	if err != nil {
		t.Fatal("failed to read blob:", err)
	}
	if data != testData {
		t.Fatal("invalid blob data")
	}
}
