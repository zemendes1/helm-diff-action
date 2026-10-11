package main

import (
	"encoding/json"
	"os"
)

type pullRequestInfo struct {
	PullRequest *struct {
		Number int `json:"number"`
	} `json:"pull_request"`
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
