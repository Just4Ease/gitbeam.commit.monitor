package scheduler

import (
	"fmt"
	"gitbeam.commit.monitor/models"
	"sync"
	"time"
)

// Job represents a task to be executed
type Job struct {
	Task   func()
	Config *models.MonitorRepositoryCommitConfig
}

func (j *Job) ID() string {
	return j.Config.ID()
}

type jobTracker struct {
	jobs      map[string]*Job
	stopChans map[string]chan bool
	mu        sync.Mutex
}

// addJob adds a new job to the scheduler
func (s *jobTracker) addJob(job *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.ID()]; exists {
		fmt.Printf("Job with ID %s already exists.\n", job.ID)
		return
	}

	s.jobs[job.ID()] = job
	stopChan := make(chan bool)
	s.stopChans[job.ID()] = stopChan
	go s.startJob(job, stopChan)
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
		s.addJob(job)
	}
}

// startJob runs the job at specified intervals
func (s *jobTracker) startJob(job *Job, stopChan chan bool) {
	ticker := time.NewTicker(60 * time.Minute * time.Duration(job.Config.DurationInHours))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			job.Task()
		case <-stopChan:
			fmt.Printf("Stopping job %s\n", job.ID)
			return
		}
	}
}
