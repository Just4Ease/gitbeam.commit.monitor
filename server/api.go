package server

import (
	"context"
	"gitbeam.baselib/utils"
	"gitbeam.commit.monitor/core"
	"gitbeam.commit.monitor/models"
	commits "gitbeam.commit.monitor/pb"
	"github.com/sirupsen/logrus"
)

type apiService struct {
	service *core.GitBeamService
	logger  *logrus.Logger
}

func (a apiService) ListCommits(ctx context.Context, params *commits.CommitFilterParams) (*commits.ListCommitResponse, error) {
	filter := models.CommitFilters{
		OwnerAndRepoName: models.OwnerAndRepoName{
			OwnerName: params.OwnerName,
			RepoName:  params.RepoName,
		},
		Limit:    params.Limit,
		Page:     params.Page,
		FromDate: nil,
		ToDate:   nil,
	}

	if params.FromDate != "" {
		filter.FromDate, _ = models.ParseDate(params.FromDate) // This will be nil if the date format doesn't work out.
	}

	if params.ToDate != "" {
		filter.ToDate, _ = models.ParseDate(params.ToDate) // This will be nil if the date format doesn't work out.
	}

	output, err := a.service.ListCommits(ctx, filter)
	if err != nil {
		return nil, err
	}

	var list []*commits.Commit
	_ = utils.UnPack(output, &list)
	return &commits.ListCommitResponse{Data: list}, nil
}

func (a apiService) GetCommitByOwnerAndSHA(ctx context.Context, params *commits.CommitByOwnerAndShaParams) (*commits.Commit, error) {
	//TODO implement me
	panic("implement me")
}

func (a apiService) ListTopCommitAuthor(ctx context.Context, params *commits.CommitFilterParams) (*commits.ListTopCommitAuthorResponse, error) {
	filter := models.CommitFilters{
		OwnerAndRepoName: models.OwnerAndRepoName{
			OwnerName: params.OwnerName,
			RepoName:  params.RepoName,
		},
		Limit:    params.Limit,
		Page:     params.Page,
		FromDate: nil,
		ToDate:   nil,
	}

	if params.FromDate != "" {
		filter.FromDate, _ = models.ParseDate(params.FromDate) // This will be nil if the date format doesn't work out.
	}

	if params.ToDate != "" {
		filter.ToDate, _ = models.ParseDate(params.ToDate) // This will be nil if the date format doesn't work out.
	}

	output, err := a.service.GetTopCommitAuthors(ctx, filter)
	if err != nil {
		return nil, err
	}

	var list []*commits.TopCommitAuthor
	_ = utils.UnPack(output, &list)
	return &commits.ListTopCommitAuthorResponse{Data: list}, nil
}

func (a apiService) HealthCheck(ctx context.Context, void *commits.Void) (*commits.HealthCheckResponse, error) {
	return &commits.HealthCheckResponse{Code: 200}, nil
}

func NewApiService(core *core.GitBeamService, logger *logrus.Logger) commits.GitBeamCommitsServiceServer {
	return &apiService{
		service: core,
		logger:  logger,
	}
}
