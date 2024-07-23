package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"gitbeam.commit.monitor/core"
	"gitbeam.commit.monitor/events/topics"
	"gitbeam.commit.monitor/models"
	"gitbeam.commit.monitor/repository"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	ErrFailedToStartMonitoringRepoCommits = errors.New("failed to start monitoring repo commits")
	ErrFailedToStopMonitoringRepoCommits  = errors.New("failed to stop monitoring repo commits")
)

// Scheduler manages the scheduling of jobs
type Scheduler struct {
	dataStore   repository.CronServiceStore
	coreService *core.GitBeamService
	logger      *logrus.Logger
	jobTracker  *jobTracker
}

// NewScheduler creates a new Scheduler
func NewScheduler(coreService *core.GitBeamService, dataStore repository.CronServiceStore, logger *logrus.Logger) *Scheduler {
	return &Scheduler{
		dataStore:   dataStore,
		coreService: coreService,
		logger:      logger.WithField("component", "scheduler").Logger,
		jobTracker: &jobTracker{
			jobs:      make(map[string]*Job),
			stopChans: make(map[string]chan bool),
		},
	}
}

func (s *Scheduler) StartMirroringRepoCommits(ctx context.Context, payload models.MonitorRepositoryCommitConfig) error {
	useLogger := s.logger.WithContext(ctx).WithField("methodName", "StartMirroringRepoCommits")

	if err := s.dataStore.SaveMonitorConfigs(ctx, payload); err != nil {
		useLogger.WithError(err).Error("Failed to save cron task in cronStore.")
		return ErrFailedToStartMonitoringRepoCommits
	}

	eventStore := s.coreService.GetEventStore()

	data, _ := json.Marshal(payload)
	_ = eventStore.Publish(topics.MonitorTaskCreated, data)
	return nil
}

func (s *Scheduler) StopMirroringRepoCommits(ctx context.Context, name models.OwnerAndRepoName) error {
	useLogger := s.logger.WithContext(ctx).WithField("methodName", "StopMirroringRepoCommits")
	existingConfig, _ := s.dataStore.GetMonitorConfig(ctx, name)
	if existingConfig == nil {
		return nil
	}

	err := s.dataStore.DeleteMonitorConfig(ctx, name)
	if err != nil {
		useLogger.WithError(err).Error("Failed to delete cron task from cronStore.")
		return ErrFailedToStopMonitoringRepoCommits
	}

	s.jobTracker.removeJob(existingConfig.ID())

	// In a real world, this deleted data would have been emitted to the eventStore for use.
	store := s.coreService.GetEventStore()

	data, _ := json.Marshal(existingConfig)
	_ = store.Publish(topics.MonitorTaskDeleted, data)
	return nil
}

func (s *Scheduler) StartScheduler() {
	s.loadExistingConfig()
	s.logger.Info("Started commit monitor scheduler...")

	<-make(chan bool)
}

func (s *Scheduler) loadExistingConfig() {
	list, err := s.dataStore.ListMonitorConfig(context.Background())
	if err != nil {
		return
	}

	for _, config := range list {
		s.jobTracker.addJob(&Job{
			Task: func() {
				ctx := context.Background()
				name := models.OwnerAndRepoName{
					OwnerName: config.OwnerName,
					RepoName:  config.RepoName,
				}

				filters := models.CommitFilters{
					OwnerAndRepoName: name,
				}

				filters.ToDate, _ = models.ParseDate(time.Now().Format(time.DateOnly))
				if lastCommit, _ := s.coreService.GetLastCommit(ctx, name); lastCommit != nil {
					filters.FromDate, _ = models.ParseDate(lastCommit.Date)
				} else {
					filters.FromDate = nil
					filters.ToDate = nil
				}

				_ = s.coreService.FetchAndSaveCommits(context.Background(), filters)
			},
			Config: config,
		})
	}
}
