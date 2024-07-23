package repository

import (
	"context"
	"gitbeam.commit.monitor/models"
	"time"
)

//go:generate mockgen -source=repository.go -destination=../mocks/data_store_mock.go -package=mocks
type DataStore interface {
	SaveCommit(ctx context.Context, payload *models.Commit) error
	ListCommits(ctx context.Context, filter models.CommitFilters) ([]*models.Commit, error)
	GetLastCommit(ctx context.Context, owner *models.OwnerAndRepoName, startTime *time.Time) (*models.Commit, error)
	GetCommitBySHA(ctx context.Context, owner models.OwnerAndRepoName, sha string) (*models.Commit, error)
	GetTopCommitAuthors(ctx context.Context, filter models.CommitFilters) ([]*models.TopCommitAuthor, error)
}

type CronServiceStore interface {
	SaveMonitorConfigs(ctx context.Context, task models.MonitorRepositoryCommitConfig) error
	ListMonitorConfig(ctx context.Context) ([]*models.MonitorRepositoryCommitConfig, error)
	GetMonitorConfig(ctx context.Context, owner models.OwnerAndRepoName) (*models.MonitorRepositoryCommitConfig, error)
	DeleteMonitorConfig(ctx context.Context, owner models.OwnerAndRepoName) error
}
