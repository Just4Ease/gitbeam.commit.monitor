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
	"sync"
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

	name := models.OwnerAndRepoName{
		OwnerName: payload.OwnerName,
		RepoName:  payload.RepoName,
	}

	existingConfig, _ := s.dataStore.GetMonitorConfig(ctx, name)
	if existingConfig != nil {
		err := s.dataStore.DeleteMonitorConfig(ctx, name)
		if err != nil {
			useLogger.WithError(err).Error("Failed to delete cron task from cronStore.")
			return ErrFailedToStopMonitoringRepoCommits
		}
		s.jobTracker.removeJob(existingConfig.ID())
	}

	if err := s.dataStore.SaveMonitorConfigs(ctx, payload); err != nil {
		useLogger.WithError(err).Error("Failed to save cron task in cronStore.")
		return ErrFailedToStartMonitoringRepoCommits
	}

	job := newJob(s.coreService, &payload)
	s.jobTracker.addJob(job)

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

	wg := sync.WaitGroup{}
	for _, config := range list {
		wg.Add(1)
		job := newJob(s.coreService, config)
		go func(job *Job) {
			defer wg.Done()
			job.Task(false)
		}(&job)
		s.jobTracker.addJob(job)
	}

	wg.Wait()
}
