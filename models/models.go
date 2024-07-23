package models

import (
	"time"
)

type Commit struct {
	Date            time.Time `json:"date"`
	Message         string    `json:"message"`
	Author          string    `json:"author"`
	RepoName        string    `json:"repoName"`
	OwnerName       string    `json:"ownerName"`
	URL             string    `json:"url"`
	SHA             string    `json:"sha"`
	ParentCommitIDs []string  `json:"parentCommitIDs"`
}

type CommitFilters struct {
	OwnerAndRepoName `json:",inline" schema:",inline"`
	Limit            int64 `json:"limit" schema:"limit,omitempty"`
	Page             int64 `json:"page" schema:"page,omitempty"`
	FromDate         *Date `json:"fromDate" schema:"fromDate,omitempty"`
	ToDate           *Date `json:"toDate" schema:"toDate,omitempty"`
}

type TopCommitAuthor struct {
	Author      string `json:"author"`
	CommitCount int    `json:"commitCount"`
}
