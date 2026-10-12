package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type pullRequestInfo struct {
	PullRequest *struct {
		Number int `json:"number"`
	} `json:"pull_request"`
}

type prCommentBody struct {
	Body string `json:"body"`
}

func getPrNumber() (int, error) {
	data, err := os.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		return 0, err
	}

	var info pullRequestInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return 0, err
	}

	if info.PullRequest == nil {
		return 0, nil
	}

	return info.PullRequest.Number, err
}

func postPRComment(comment string, prNumber int, githubAPIURL string, githubRepository string, githubToken string) error {
	endpoint := fmt.Sprintf("%s/repos/%s/issues/%d/comments", githubAPIURL, githubRepository, prNumber)

	commentBody, err := json.Marshal(prCommentBody{Body: comment})
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(commentBody))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubToken))
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusCreated {
		return err
	}

	defer response.Body.Close()

	return nil
}
