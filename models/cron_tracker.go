package models

import "fmt"

type MonitorRepositoryCommitConfig struct {
	RepoName        string `json:"repoName"`
	OwnerName       string `json:"ownerName"`
	FromDate        string `json:"fromDate"`
	ToDate          string `json:"toDate"`
	DurationInHours int    `json:"durationInHours"`
}

func (c *MonitorRepositoryCommitConfig) ID() string {
	return fmt.Sprintf("%s/%s", c.RepoName, c.OwnerName)
}
