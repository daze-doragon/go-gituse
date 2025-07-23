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

# Test
This project includes tests for the `go-gituse` library.
You can run the tests from the root of the project with the following command:
```bash
go test ./test
```

## Test Environment
To run the tests, you must create a `test.env` file in the `test` directory to provide necessary environment variables.
```env
TEST_GIT_REPOSITORY=test_repository
TEST_GIT_OWNER=owner_of_test_repository
TEST_GIT_TOKEN=fine_grained_token
TEST_GIT_AUTHOR=author_for_test_commit
TEST_GIT_EMAIL=email_for_test_commit
```

## About the Fine-Grained Token for Testing
The token used for testing must have the following permissions:
- **Repository access**: All repositories (required to create and delete repositories).
- **Content permissions**: Read and write access (required for commit creation and reading refs).
- **Administration permissions**: Read and write access (required to create and delete repositories).

## Test Details
The test performs the following steps:
1. **Create a private repository.**  
    If repository creation fails, the test will terminate immediately.
2. **Test CreateCommitByLocalDir method.**  
    This creates a commit using files in the ./test/repo_test directory.
3. **Test MakeCommitElementByFileData method.**  
    It tests creating commit elements from both a string-based file content and an existing .png file.
5. **Test CreateCommitByElement method.**  
6. **Delete the private repository.**
