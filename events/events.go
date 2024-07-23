package events

import (
	"context"
	"gitbeam.baselib/store"
	"gitbeam.baselib/utils"
	"gitbeam.commit.monitor/core"
	"gitbeam.commit.monitor/events/topics"
	"gitbeam.commit.monitor/models"
	"github.com/sirupsen/logrus"
)

type EventHandlers struct {
	logger        *logrus.Logger
	service       *core.GitBeamService
	eventStore    store.EventStore
	subscriptions []func() error
}

func NewEventHandler(
	eventStore store.EventStore,
	logger *logrus.Logger,
	service *core.GitBeamService,
) EventHandlers {
	return EventHandlers{
		logger:     logger.WithField("module", "EventHandler").Logger,
		service:    service,
		eventStore: eventStore,
	}
}

func (e EventHandlers) Listen() {
	useLogger := e.logger.WithField("methodName", "Listen")
	e.subscriptions = append(
		e.subscriptions,
		e.handleCronTaskCreated,
	)

	for _, sub := range e.subscriptions {
		if err := sub(); err != nil {
			useLogger.WithError(err).Fatal("failed to mount subscription")
		}
	}

	<-make(chan bool)
}

func (e EventHandlers) handleCronTaskCreated() error {
	return e.eventStore.Subscribe(topics.MonitorTaskCreated, func(event store.Event) error {
		e.logger.Infof("received event on %s", topics.MonitorTaskCreated)

		var config models.MonitorRepositoryCommitConfig
		_ = utils.UnPack(event.Data(), &config)
		ctx := context.Background()

		params := models.CommitFilters{
			OwnerAndRepoName: models.OwnerAndRepoName{
				OwnerName: config.OwnerName,
				RepoName:  config.RepoName,
			},
			FromDate: nil,
			ToDate:   nil,
		}

		if config.FromDate != "" {
			params.FromDate, _ = models.ParseDate(config.FromDate) // Defaults to null if nothing.
		}

		if config.ToDate != "" {
			params.ToDate, _ = models.ParseDate(config.ToDate) // Defaults to null if nothing.
		}

		return e.service.FetchAndSaveCommits(ctx, params)
	})
}
