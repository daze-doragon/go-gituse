package githubapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type GitClient struct {
	Token      string
	Owner      string
	Repository string
	Branch     string
}

// GetRef APIの結果を受け取る構造体
type RefResponse struct {
	Ref     string `json:"ref"`
	Node_id string `json:"node_id"`
	Url     string `json:"url"`
	Object  struct {
		Sha  string `json:"sha"` //latest commit sha.
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"object"`
}

// GetCommit APIの結果を受け取る構造体
type CommitResponse struct {
	Sha      string `json:"sha"`
	Url      string `json:"url"`
	Html_url string `json:"html_url"`
	Author   struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Date  string `json:"date"`
	} `json:"author"`
	Committer struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Date  string `json:"date"`
	} `json:"committer"`
	Tree struct {
		Sha string `json:"sha"`
		Url string `json:"url"`
	} `json:"tree"`
	Message string `json:"message"`
	Parents []struct {
		Sha      string `json:"sha"`
		Url      string `json:"url"`
		Html_url string `json:"html_url"`
	} `json:"parents"`
}

// CreateBlob APIの結果を受け取る構造体
type CreateBlobResponse struct {
	Sha string `json:"sha"`
	Url string `json:"url"`
}

// CreateBlob APIのbodyに指定する構造体
type BlobData struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// CreaateTree  APIの結果を受け取る構造体
type CreateTreeResponse struct {
	SHA  string `json:"sha"`
	URL  string `json:"url"`
	Tree []struct {
		Path string `json:"path"`
		Mode string `json:"mode"`
		Type string `json:"type"`
		SHA  string `json:"sha"`
		Size int    `json:"size"`
		URL  string `json:"url"`
	} `json:"tree"`
	Truncated bool `json:"truncated"`
}

// CreateTree APIのbodyに指定する構造体
type TreeData struct {
	Base_tree *string            `json:"base_tree,omitempty"`
	Tree      []*TreeDataElement `json:"tree"`
}

// CreateTree APIのbodyに指定するデータの内、Tree要素を表現する構造体
type TreeDataElement struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	Sha  string `json:"sha"`
}

// CreaateCommit APIの結果を受け取る構造体
type CreateCommitResponse struct {
	Sha      string `json:"sha"`
	Url      string `json:"url"`
	Html_url string `json:"html_url"`
	Author   struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Date  string `json:"date"`
	} `json:"author"`
	Commiter struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Date  string `json:"date"`
	} `json:"commiter"`
	Tree struct {
		Sha string `json:"sha"`
		Url string `json:"url"`
	} `json:"tree"`
	Message string `json:"message"`
	Parents []struct {
		Sha      string `json:"sha"`
		Url      string `json:"url"`
		Html_url string `json:"html_url"`
	} `json:"parents"`
}

// CreateCommit APIのbodyに指定する構造体
type CommitData struct {
	Message string `json:"message"`
	Author  *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Date  string `json:"date"`
	} `json:"author"`
	Parents []string `json:"parents"`
	Tree    string   `json:"tree"`
}

// UpdateRef APIのbodyに指定する構造体
type UpdRefData struct {
	Sha   string `json:"sha"`
	Force bool   `json:"force"`
}

// CreateRef APIのbodyに指定する構造体
type CreateRefData struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

// UpdateRef  APIの結果を受け取る構造体
type UpdateRefResponse struct {
	Ref     string `json:"ref"`
	Node_id string `json:"node_id"`
	Url     string `json:"url"`
	Object  struct {
		Sha  string `json:"sha"`
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"object"`
}

// github apiコール用のクライアントを取得する
func GetGitClient(token *string, owner string, repoName string, branch *string) (*GitClient, error) {
	if token == nil {
		s := ""
		token = &s
	}
	if branch == nil {
		s := "main"
		branch = &s
	}

	client := &GitClient{
		Token:      *token,
		Owner:      owner,
		Repository: repoName,
		Branch:     *branch,
	}
	return client, nil
}

func (git *GitClient) GetLatestCommitSha() (string, error) {
	ref, err := git.GetLatestRef()
	if err != nil {
		log.Printf("error occured when getting latest ref." + err.Error())
	}
	return ref.Object.Sha, nil
}

func (git *GitClient) GetLatestRef() (*RefResponse, error) {
	endPoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/%s", git.Owner, git.Repository, git.Branch)
	headerMap := make(map[string]string)
	headerMap["Authorization"] = "Bearer " + git.Token
	resp, err := requestSend("GET", endPoint, nil, headerMap)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		emptyFlg, err := git.IsEmptyRepository()
		if err != nil {
			return nil, fmt.Errorf("error occured when check empty repository. %w", err)
		}
		if emptyFlg {
			return &RefResponse{}, nil
		}
		return nil, errors.New(string(respData))
	}
	ref := &RefResponse{} //use RefResponse struct for return.
	err = json.Unmarshal(respData, ref)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func (git *GitClient) CreateBlob(blob *BlobData) (*CreateBlobResponse, error) {
	endPoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/blobs", git.Owner, git.Repository)
	headerMap := make(map[string]string)
	headerMap["Authorization"] = "Bearer " + git.Token
	bodyData, err := json.Marshal(blob)
	if err != nil {
		return nil, err
	}
	resp, err := requestSend("POST", endPoint, bytes.NewReader(bodyData), headerMap)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	blobResponse := &CreateBlobResponse{}
	err = json.Unmarshal(respData, blobResponse)
	if err != nil {
		return nil, err
	}
	return blobResponse, nil
}

func (git *GitClient) CreateTree(tree *TreeData) (*CreateTreeResponse, error) {
	endPoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees", git.Owner, git.Repository)
	headerMap := make(map[string]string)
	headerMap["Authorization"] = "Bearer " + git.Token
	bodyData, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}
	resp, err := requestSend("POST", endPoint, bytes.NewReader(bodyData), headerMap)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(string(respData))
	}
	treeResponse := &CreateTreeResponse{}
	err = json.Unmarshal(respData, treeResponse)
	if err != nil {
		return nil, err
	}
	return treeResponse, nil
}

func (git *GitClient) CreateCommit(commit *CommitData) (*CreateCommitResponse, error) {
	endPoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits", git.Owner, git.Repository)
	headerMap := make(map[string]string)
	headerMap["Authorization"] = "Bearer " + git.Token
	bodyData, err := json.Marshal(commit)
	if err != nil {
		return nil, err
	}
	resp, err := requestSend("POST", endPoint, bytes.NewReader(bodyData), headerMap)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return nil, errors.New(string(respData))
	}
	commitResponse := &CreateCommitResponse{}
	err = json.Unmarshal(respData, commitResponse)
	if err != nil {
		return nil, err
	}
	return commitResponse, nil
}

func (git *GitClient) CreateRef(refData *CreateRefData) (*UpdateRefResponse, error) {
	endPoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", git.Owner, git.Repository)
	headerMap := make(map[string]string)
	headerMap["Authorization"] = "Bearer " + git.Token
	bodyData, err := json.Marshal(refData)
	if err != nil {
		return nil, err
	}
	resp, err := requestSend("POST", endPoint, bytes.NewReader(bodyData), headerMap)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(string(respData))
	}
	createRefResponse := &UpdateRefResponse{}
	err = json.Unmarshal(respData, createRefResponse)
	if err != nil {
		return nil, err
	}
	return createRefResponse, nil
}

func (git *GitClient) UpdateRef(refData *UpdRefData) (*UpdateRefResponse, error) {
	endPoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/%s", git.Owner, git.Repository, git.Branch)
	headerMap := make(map[string]string)
	headerMap["Authorization"] = "Bearer " + git.Token
	bodyData, err := json.Marshal(refData)
	if err != nil {
		return nil, err
	}
	resp, err := requestSend("PATCH", endPoint, bytes.NewReader(bodyData), headerMap)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(string(respData))
	}
	updRefResponse := &UpdateRefResponse{}
	err = json.Unmarshal(respData, updRefResponse)
	if err != nil {
		return nil, err
	}
	return updRefResponse, nil
}

func (git *GitClient) GetCommit(commitId string) (*CommitResponse, error) {
	endPoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits/%s", git.Owner, git.Repository, commitId)
	headerMap := make(map[string]string)
	headerMap["Authorization"] = "Bearer " + git.Token
	resp, err := requestSend("GET", endPoint, nil, headerMap)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	cmtResponse := &CommitResponse{}
	err = json.Unmarshal(respData, cmtResponse)
	if err != nil {
		return nil, err
	}
	return cmtResponse, nil
}

// リポジトリが空の状態かを確認する(未テスト 使えるかわからない)
func (git *GitClient) IsEmptyRepository() (bool, error) {
	endPoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/heads/%s", git.Owner, git.Repository, git.Branch)
	headerMap := make(map[string]string)
	headerMap["Authorization"] = "Bearer " + git.Token
	resp, err := requestSend("GET", endPoint, nil, headerMap)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusConflict {
		return true, nil
	} else if resp.StatusCode == http.StatusOK {
		return false, nil
	} else {
		return false, errors.New(string(respData))
	}
}

// httpリクエスト送信
func requestSend(method string, endPoint string, body io.Reader, headerMap map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, endPoint, body) //reqeuest
	if err != nil {
		return nil, err
	}
	for key, value := range headerMap {
		req.Header.Set(key, value) //header
	}
	client := &http.Client{}
	resp, err := client.Do(req) //send http
	if err != nil {
		return nil, err
	}
	return resp, nil
}
