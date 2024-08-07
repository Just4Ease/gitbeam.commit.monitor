package cron

import (
	"context"
	"encoding/json"
	"errors"
	"gitbeam.commit.monitor/core"
	"gitbeam.commit.monitor/events/topics"
	"gitbeam.commit.monitor/models"
	"gitbeam.commit.monitor/repository"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Service struct {
	logger    *logrus.Logger
	service   *core.GitBeamService
	scheduler gocron.Scheduler
	cronStore repository.CronServiceStore
}

var (
	ErrFailedToStartMirroringRepoCommits = errors.New("failed to start mirroring repo commits")
	ErrFailedToStopMirroringRepoCommits  = errors.New("failed to stop mirroring repo commits")
)

func NewCronService(cronStore repository.CronServiceStore, service *core.GitBeamService, logger *logrus.Logger) *Service {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logger.WithError(err).Fatal("Failed to start cron service")
	}

	return &Service{
		logger:    logger.WithField("moduleName", "CronService").Logger,
		service:   service,
		scheduler: scheduler,
		cronStore: cronStore,
	}
}

func (s Service) Start() {
	// Schedule the task to run every 10 minutes
	_, err := s.scheduler.NewJob(
		gocron.DurationJob(time.Minute*10),
		gocron.NewTask(s.listCronTasksAndExecuteRepoCommitsMirroring),
	)
	if err != nil {
		s.logger.WithError(err).Fatal("failed to start job")
	}

	s.scheduler.Start() // This is non-blocking.
	<-make(chan bool)   // use this to block and hold the cron service.
}

func (s Service) StartMirroringRepoCommits(ctx context.Context, payload models.MirrorRepoCommitsRequest) (*models.Repo, error) {
	useLogger := s.logger.WithContext(ctx).WithField("methodName", "StartMirroringRepoCommits")
	repo, err := s.service.GetByOwnerAndRepoName(ctx, &models.OwnerAndRepoName{
		OwnerName: payload.OwnerName,
		RepoName:  payload.RepoName,
	})
	if err != nil {
		useLogger.WithError(err).Error("error attempting to beam repository commits.")
		return nil, err
	}

	if repo.Meta == nil {
		repo.Meta = make(map[string]any)
	}

	var fromDate *time.Time
	var toDate *time.Time

	if payload.FromDate != nil {
		fromDate = &payload.FromDate.Time
		repo.Meta["fromDate"] = payload.FromDate.String()
	}

	if payload.ToDate != nil {
		toDate = &payload.ToDate.Time
		repo.Meta["toDate"] = payload.ToDate.String()
	}

	err = s.cronStore.SaveMonitorConfigs(ctx, models.MonitorRepositoryCommitConfig{
		RepoName:  repo.Name,
		OwnerName: repo.Owner,
		FromDate:  fromDate,
		ToDate:    toDate,
	})
	if err != nil {
		useLogger.WithError(err).Error("Failed to save cron task in cronStore.")
		return nil, ErrFailedToStartMirroringRepoCommits
	}

	eventStore := s.service.GetEventStore()

	data, _ := json.Marshal(repo)
	_ = eventStore.Publish(topics.CronTaskCreated, data)
	return repo, nil
}

func (s Service) StopMirroringRepoCommits(ctx context.Context, name models.OwnerAndRepoName) error {
	useLogger := s.logger.WithContext(ctx).WithField("methodName", "StopMirroringRepoCommits")
	err := s.cronStore.DeleteMonitorConfig(ctx, name)
	if err != nil {
		useLogger.WithError(err).Error("Failed to delete cron task from cronStore.")
		return ErrFailedToStopMirroringRepoCommits
	}

	// In a real world, this deleted data would have been emitted to the eventStore for use.
	return nil
}

func (s Service) listCronTasksAndExecuteRepoCommitsMirroring() {
	useLogger := s.logger.WithField("methodName", "listCronTasksAndExecuteRepoCommitsMirroring")
	useLogger.Info("Started fetching and saving commits")
	ctx := context.Background()
	trackers, _ := s.cronStore.ListMonitorConfig(ctx)
	wg := &sync.WaitGroup{}
	for _, tracker := range trackers {
		wg.Add(1)
		go func(name models.OwnerAndRepoName) {
			defer wg.Done()

			filters := models.CommitFilters{
				OwnerAndRepoName: name,
			}

			filters.ToDate, _ = models.ParseDate(time.Now().Format(time.DateOnly))
			if lastCommit, _ := s.service.GetLastCommit(ctx, name); lastCommit != nil {
				filters.FromDate, _ = models.ParseDate(lastCommit.Date.Format(time.DateTime))
			} else {
				filters.FromDate = nil
				filters.ToDate = nil
			}

			_ = s.service.FetchAndSaveCommits(ctx, filters)
		}(models.OwnerAndRepoName{
			OwnerName: tracker.OwnerName,
			RepoName:  tracker.RepoName,
		})
	}
	wg.Wait()
	useLogger.Info("Finished fetching and saving commits")
}
