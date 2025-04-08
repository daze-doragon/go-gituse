# introduction
Using Github API, Create Commit with code.
for instance.
  You can create commit with file that created dynamically in your program.
  You can create commit with local directory that simulate github repository.
You need to get fine-grained access token for github repository.

# Installation
```sh
go get github.com/daze-doragon/go-gituse
```

# Usage
1. get gitInfo object → CreateCommitByLocalDir
  In this usage, commits are created by specifying a directory that mimics the structure of a local GitHub repository.
2. get gitInfo object → Make CommitElement → CreateCommitByCommitElements
   In this usage, commits are created with CommitElements. CommitElement is a variable that represents changes at the file level when creating a commit.

# Sample Code
```go
package main

import (
	"fmt"

	service "github.com/daze-doragon/go-gituse/pkg/service"
)

func main() {
	token := "your github Fine-grained personal access tokens"
	owner := "your github user name"
	repo := "your github repository"
	branch := "your branch"
	author := "committer author info"
	email := "committer email address"

	// Initialize a Git info object.
	// If nil is used for the branch, gitInfo will default to the 'main' branch.
	// For details about the token, refer to "https://github.blog/security/application-security/introducing-fine-grained-personal-access-tokens-for-github"
	gitInfo, err := service.GetGitInfo(&token, owner, repo, &branch, author, email)
	if err != nil {
		// Error occurred when getting Git info.
	}

	// Usage of MakeCommitElementByFileData
	// Below, cmtElement1 represents a staged file "files/file_no1.txt" with its content updated to the fileContents variable.
	// If "files/file_no1.txt" already exists in your repository, its content will be updated.
	// If it does not exist, the file will be added.
	fileContents := "It is content of file_no1.txt"
	cmtElement1, err := service.MakeCommitElementByFileData("files/file_no1.txt", fileContents, service.Utf8)
	if err != nil {
		// Error occurred when creating commitElement.
	}

	// Usage of CreateCommitByCommitElements
	// Create a commit with a list of CommitElements.
	// Each CommitElement represents a staged file.
	cmtElementList := []*service.CommitElement{cmtElement1}
	cmtResp1, err := gitInfo.CreateCommitByElement("commit by CreateCommitByElement.", cmtElementList)
	if err != nil {
		// Error occurred when creating the commit.
	}
	fmt.Print(cmtResp1.Sha) // Print commit SHA.

	// CreateCommitByLocalDir
	// Create a commit using local files.
	// This method automatically creates CommitElements for files in "local_files".
	// It only handles "add" and "update", but not "delete".
	// If your repository has "file_no3.txt" but it does not exist in "local_files", "file_no3.txt" will not be deleted.
	cmtResp2, err := gitInfo.CreateCommitByLocalDir("commit by CreateCommitByLocalDir", "./local_files")
	if err != nil {
		// Error occurred when creating the commit.
	}
	fmt.Print(cmtResp2.Sha) // Print commit SHA.
}
```
