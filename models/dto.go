package models

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

type Result struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Success bool   `json:"success"`
}

type OwnerAndRepoName struct {
	OwnerName string `json:"ownerName" schema:"ownerName"`
	RepoName  string `json:"repoName" schema:"repoName"`
}

type MirrorRepoCommitsRequest struct {
	FromDate         *Date `json:"fromDate,omitempty"`
	ToDate           *Date `json:"toDate,omitempty"`
	OwnerAndRepoName `json:",inline"`
}

func (s OwnerAndRepoName) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.OwnerName, validation.Required),
		validation.Field(&s.RepoName, validation.Required))
}
