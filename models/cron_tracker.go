package models

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
)

type MonitorRepositoryCommitConfig struct {
	RepoName        string `json:"repoName"`
	OwnerName       string `json:"ownerName"`
	FromDate        string `json:"fromDate"`
	ToDate          string `json:"toDate"`
	DurationInHours int64  `json:"durationInHours"`
}

func (c MonitorRepositoryCommitConfig) ID() string {
	return fmt.Sprintf("%s/%s", c.RepoName, c.OwnerName)
}

func (c MonitorRepositoryCommitConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.OwnerName, validation.Required),
		validation.Field(&c.RepoName, validation.Required),
		validation.Field(&c.DurationInHours, validation.Required, validation.Min(1)),
	)
}
