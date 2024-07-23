package models

type CronTask struct {
	RepoName  string `json:"repoName"`
	OwnerName string `json:"ownerName"`
	FromDate  string `json:"fromDate"`
	ToDate    string `json:"toDate"`
}
