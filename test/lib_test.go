package test

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/joho/godotenv"

	service "github.com/daze-doragon/go-gituse/pkg/service"
)

func TestLib(t *testing.T) {
	err := godotenv.Load("./test.env")
	if err != nil {
		t.Error("Failed to read ./test.env.", err)
	}
	token := os.Getenv("TEST_GIT_TOKEN")
	owner := os.Getenv("TEST_GIT_OWNER")
	repository := os.Getenv("TEST_GIT_REPOSITORY")
	author := os.Getenv("TEST_GIT_AUTHOR")
	email := os.Getenv("TEST_GIT_EMAIL")
	git, _ := service.GetGitInfo(&token, owner, repository, nil, author, email)

	err = git.CreatePrivateRepo()
	if err != nil {
		t.Fatal(err)
	}

	_, err = git.CreateCommitByLocalDir("commit by CreateCommitByLocalDir.", "./repo_test")
	if err != nil {
		t.Error(err)
	}

	var eleList []*service.CommitElement
	fileContent := "This is test file for MakeCommitElementByFileData method."
	ele, err := service.MakeCommitElementByFileData("test2.txt", fileContent, service.Utf8)
	if err != nil {
		t.Error(err)
	}
	eleList = append(eleList, ele)
	fileData, err := os.ReadFile("./test_data/test3.png")
	if err != nil {
		t.Error(err)
	}
	ele, err = service.MakeCommitElementByFileData("test3.png", base64.StdEncoding.EncodeToString(fileData), service.FormattedBinary)
	if err != nil {
		t.Error(err)
	}
	eleList = append(eleList, ele)

	_, err = git.CreateCommitByElement("commit by CreateCommitByElement.", eleList)
	if err != nil {
		t.Error(err)
	}

	err = git.DeletePrivateRepo()
	if err != nil {
		t.Log("Error occured when delete test repository. you should check if the test repository is deleted.")
		t.Fatal(err)
	}
}
