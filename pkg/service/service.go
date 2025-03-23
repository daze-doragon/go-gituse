package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	githubapi "github.com/daze-doragon/go-gituse/pkg/githubapi"
)

const (
	FormattedBinary int = 1
	Utf8            int = 2
)

type GitInfo struct {
	client       *githubapi.GitClient
	author_name  string
	author_email string
}

// コミットの要素になるデータ(blob単位)
type CommitElement struct {
	pathInRepo   string
	pathInLocal  string
	content      string
	encodingType int //0:データなし 1:base64_binary 2:utf-8
	blobSha      string
}

// GitHub操作用のオブジェクト
func GetGitInfo(token *string, owner string, repoName string, branch *string, author string, email string) (*GitInfo, error) {
	if token == nil {
		s := ""
		token = &s
	}
	if branch == nil {
		s := "main"
		branch = &s
	}

	gitClient := &githubapi.GitClient{
		Token:      *token,
		Owner:      owner,
		Repository: repoName,
		Branch:     *branch,
	}
	gitInfo := &GitInfo{
		client:       gitClient,
		author_name:  author,
		author_email: email,
	}
	return gitInfo, nil
}

// ローカルのファイルパスを指定してCommitElementを作成する。
func MakeCommitElementListByLocalPath(localPath string) ([]*CommitElement, error) {
	var commitEleList []*CommitElement
	err := filepath.WalkDir(localPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			separator := string(os.PathSeparator)
			pathEle := strings.Split(path, separator)
			repoPath := strings.Join(pathEle[1:], "/")
			ele := &CommitElement{
				pathInRepo:  repoPath,
				pathInLocal: path,
			}
			commitEleList = append(commitEleList, ele)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error occured when process file in localPath. %w", err)
	}
	return commitEleList, nil
}

// ファイルデータを指定してCommitElementを作成する。encType 1:binary(base64encode) 2:utf-8
func MakeCommitElementByFileData(repoPath string, fileContent string, encType int) (*CommitElement, error) {
	if !(encType == FormattedBinary || encType == Utf8) {
		return nil, errors.New("invalid encType 1:binary(base64encode) 2:utf-8")
	}
	if encType == FormattedBinary {
		if !isBase64(fileContent) {
			return nil, errors.New("invalid base64 data.")
		}
	} else {
		if !utf8.Valid([]byte(fileContent)) {
			return nil, errors.New("invalid utf8 data.")
		}
	}
	element := &CommitElement{
		pathInRepo:   repoPath,
		content:      fileContent,
		encodingType: encType,
	}
	return element, nil
}

// 空のリポジトリかを確認する(未テスト)
func (gitInfo *GitInfo) IsEmptyRepository() (bool, error) {
	git, err := githubapi.GetGitClient(&gitInfo.client.Token, gitInfo.client.Owner, gitInfo.client.Repository, &gitInfo.client.Branch)
	if err != nil {
		return false, err
	}
	return git.IsEmptyRepository()
}

// ローカルのパスを指定しコミットを作る。指定したパスはリポジトリのルートと認識しそれに応じたパスでコミットを作成する。
func (gitInfo *GitInfo) CreateCommitByLocalDir(commitMsg string, localPath string) (*githubapi.CreateCommitResponse, error) {
	commitEleList, err := MakeCommitElementListByLocalPath(localPath)
	if err != nil {
		return nil, fmt.Errorf("error occured when make commitElementList. %w", err)
	}
	createCommitResp, err := gitInfo.CreateCommitByElement(commitMsg, commitEleList)
	if err != nil {
		return nil, fmt.Errorf("error occured when createCommit. %w", err)
	}
	return createCommitResp, nil
}

// 作成したCommitElement配列を指定してコミットを作成する。戻り値はCreateCommit APIのレスポンス構造体 CommitIDにはcoreateCommitResponse.Shaでアクセスできる。
func (gitInfo *GitInfo) CreateCommitByElement(commitMsg string, elementList []*CommitElement) (*githubapi.CreateCommitResponse, error) {
	git := gitInfo.client
	var isEmptyRepo bool

	//refを取得して最新commitを確認
	ref, err := git.GetLatestRef()
	if err != nil {
		return nil, fmt.Errorf("error occured when get the latest ref. %w", err)
	}
	isEmptyRepo = (ref.Ref == "")

	//最新commitを取得してbasetree取得
	var commitResp *githubapi.CommitResponse
	if !isEmptyRepo {
		commitResp, err = git.GetCommit(ref.Object.Sha)
		if err != nil {
			return nil, fmt.Errorf("error occured when get the latest commit. %q", err)
		}
	} else {
		commitResp = &githubapi.CommitResponse{}
	}

	//CommitElementをループしてblobを作成
	var treeDataEleList []*githubapi.TreeDataElement
	for _, element := range elementList {
		blobData, err := getBlobDataByElement(element)
		if err != nil {
			return nil, fmt.Errorf("error occured when create blobData. %w", err)
		}
		createBlobResp, err := git.CreateBlob(blobData)
		if err != nil {
			return nil, fmt.Errorf("error occured when create blob %w", err)
		}
		element.blobSha = createBlobResp.Sha //blob id 更新

		treeDataEle := &githubapi.TreeDataElement{
			Path: element.pathInRepo,
			Mode: "100644",
			Type: "blob",
			Sha:  element.blobSha,
		}
		treeDataEleList = append(treeDataEleList, treeDataEle)
	}

	//作成したblobをまとめるtreeを作成
	var baseTree *string
	if isEmptyRepo {
		baseTree = nil
	} else {
		baseTree = &commitResp.Tree.Sha
	}
	tree := &githubapi.TreeData{
		Base_tree: baseTree,
		Tree:      treeDataEleList,
	}
	createTreeResp, err := git.CreateTree(tree)
	if err != nil {
		return nil, fmt.Errorf("error occured when create tree. %w", err)
	}

	//commit:date
	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Tokyo")

	//commitを作成
	var parents []string
	if !isEmptyRepo {
		parents = append(parents, commitResp.Sha)
	}
	commitData := &githubapi.CommitData{
		Message: commitMsg,
		Author: &struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Date  string `json:"date"`
		}{
			Name:  gitInfo.author_name,
			Email: gitInfo.author_email,
			Date:  now.In(loc).Format("2006-01-02T15:04:05+09:00"),
		},
		Parents: parents,
		Tree:    createTreeResp.SHA,
	}
	createCommitResp, err := git.CreateCommit(commitData)
	if err != nil {
		return nil, fmt.Errorf("error occured when CreateCommit. %w", err)
	}

	//refの更新
	if isEmptyRepo {
		//空のリポジトリだった場合はref作成
		refData := &githubapi.CreateRefData{
			Ref: fmt.Sprintf("ref/head/%s", git.Branch),
			Sha: createCommitResp.Sha,
		}
		updateRefResp, err := git.CreateRef(refData)
		if err != nil {
			return nil, fmt.Errorf("error occured when CreateRef. %w", err)
		}
		fmt.Printf(updateRefResp.Ref)
	} else {
		//これ以前のコミットがある場合はref更新
		refData := &githubapi.UpdRefData{
			Sha:   createCommitResp.Sha,
			Force: false,
		}
		updRefResp, err := git.UpdateRef(refData)
		if err != nil {
			return nil, fmt.Errorf("error occured when UpdateRef. %w", err)
		}
		fmt.Printf(updRefResp.Ref)
	}
	return createCommitResp, nil
}

// 文字列がbase64エンコードされたものかを確認する。
func isBase64(s string) bool {
	if len(s)%4 != 0 {
		return false
	}
	base64Rgx := regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
	if !base64Rgx.MatchString(s) {
		return false
	}
	decode, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return false
	}
	return utf8.Valid(decode)
}

// CommitElementからBlobData構造体を返す。elementがローカルパスを持つ場合は、対象ファイルのcontentをbinaryでBlobを作る。
func getBlobDataByElement(element *CommitElement) (*githubapi.BlobData, error) {
	var blobData *githubapi.BlobData
	if element.pathInLocal != "" {
		//ローカルパス指定
		file, err := os.Open(element.pathInLocal)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		var data []byte
		buf := make([]byte, 4096)
		for {
			n, err := file.Read(buf)
			if err == io.EOF {
				break
			}
			data = append(data, buf[:n]...)
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		blobData = &githubapi.BlobData{
			Content:  encoded,
			Encoding: "base64",
		}
	} else if element.content != "" && element.encodingType != 0 {
		//content指定
		var encStr string
		if element.encodingType == FormattedBinary {
			encStr = "base64"
		} else if element.encodingType == Utf8 {
			encStr = "utf8"
		} else {
			return nil, errors.New("elementのencodingが不正です。")
		}
		blobData = &githubapi.BlobData{
			Content:  element.content,
			Encoding: encStr,
		}
	}
	return blobData, nil
}
