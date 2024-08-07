package scheduler

import (
	"context"
	"fmt"
	"gitbeam.commit.monitor/core"
	"gitbeam.commit.monitor/models"
	"sync"
	"time"
)

// Job represents a task to be executed
type Job struct {
	Task   func(withDateRange bool)
	Config *models.MonitorRepositoryCommitConfig
}

func newJob(coreService *core.GitBeamService, cfg *models.MonitorRepositoryCommitConfig) Job {
	return Job{
		Config: cfg,
		Task: func(withDateRange bool) {
			ctx := context.Background()
			name := models.OwnerAndRepoName{
				OwnerName: cfg.OwnerName,
				RepoName:  cfg.RepoName,
			}

			filters := models.CommitFilters{
				OwnerAndRepoName: name,
				FromDate:         nil,
				ToDate:           nil,
				Limit:            0,
				Page:             0,
			}

			if withDateRange {
				if cfg.FromDate != "" {
					if date, _ := models.ParseDate(cfg.FromDate); date != nil {
						filters.FromDate = date
					}
				}

				if cfg.ToDate != "" {
					if date, _ := models.ParseDate(cfg.ToDate); date != nil {
						filters.ToDate = date
					}
				}
			} else {
				filters.ToDate, _ = models.ParseDate(time.Now().Format(time.DateOnly))
				if lastCommit, _ := coreService.GetLastCommit(ctx, name); lastCommit != nil {
					filters.FromDate, _ = models.ParseDate(lastCommit.Date.Format(time.DateOnly))
					filters.ToDate, _ = models.ParseDate(time.Now().Format(time.DateOnly))
				} else {
					filters.FromDate = nil
					filters.ToDate = nil
				}
			}

			_ = coreService.FetchAndSaveCommits(context.Background(), filters)
		},
	}
}

func (j Job) ID() string {
	return j.Config.ID()
}

type jobTracker struct {
	jobs      map[string]*Job
	stopChans map[string]chan bool
	mu        sync.Mutex
}

// addJob adds a new job to the scheduler
func (s *jobTracker) addJob(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.ID()]; exists {
		fmt.Printf("Job with ID %s already exists.\n", job.ID)
		return
	}

	s.jobs[job.ID()] = &job
	stopChan := make(chan bool)
	s.stopChans[job.ID()] = stopChan
	go s.startJob(&job, stopChan)
}

// removeJob removes a job from the scheduler
func (s *jobTracker) removeJob(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stopChan, exists := s.stopChans[id]; exists {
		close(stopChan)
		delete(s.stopChans, id)
	}

	delete(s.jobs, id)
}

// updateJob updates an existing job's interval
func (s *jobTracker) updateJob(cfg models.MonitorRepositoryCommitConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := cfg.ID()

	if job, exists := s.jobs[id]; exists {
		s.removeJob(id)
		s.addJob(*job)
	}
}

// startJob runs the job at specified intervals
func (s *jobTracker) startJob(job *Job, stopChan chan bool) {
	ticker := time.NewTicker(60 * time.Minute * time.Duration(job.Config.DurationInHours))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			job.Task(false)
		case <-stopChan:
			fmt.Printf("Stopping job %s\n", job.ID)
			return
		}
	}
}
