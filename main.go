package main

import (
	"fmt"
	"gitbeam.baselib/store"
	"gitbeam.commit.monitor/config"
	"gitbeam.commit.monitor/core"
	"gitbeam.commit.monitor/events"
	"gitbeam.commit.monitor/repository"
	"gitbeam.commit.monitor/repository/sqlite"
	"gitbeam.commit.monitor/scheduler"
	"gitbeam.commit.monitor/server"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	var eventStore store.EventStore
	var dataStore repository.DataStore
	var cronStore repository.CronServiceStore
	var err error

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	secrets := config.GetSecrets()

	//Using SQLite as the mini persistent storage.
	//( in a real world system, this would be any production level or vendor managed db )
	if dataStore, err = sqlite.NewSqliteRepo(secrets.CommitDatabaseName); err != nil {
		logger.WithError(err).Fatal("failed to initialize sqlite database repository for cron store.")
	}

	// A channel based pub/sub messaging system.
	//( in a real world system, this would be apache-pulsar, kafka, nats.io or rabbitmq )
	eventStore = store.NewEventStore(logger)

	// If the dependencies were more than 3, I would use a variadic function to inject them.
	//Clarity is better here for this exercise.
	coreService := core.NewGitBeamService(logger, eventStore, dataStore, nil)

	// To handle event-based background activities. ( in a real world system, this would be apache-pulsar, kafka, nats.io or rabbitmq )
	go events.NewEventHandler(eventStore, logger, coreService).Listen()

	//Using SQLite as the mini persistent storage.
	//( in a real world system, this would be any production level or vendor managed db )
	if cronStore, err = sqlite.NewSqliteCronStore("cron_store.db"); err != nil {
		logger.WithError(err).Fatal("failed to initialize sqlite database repository for cron store.")
	}

	schedulerService := scheduler.NewScheduler(coreService, cronStore, logger)
	go schedulerService.StartScheduler()

	address := fmt.Sprintf("0.0.0.0:%s", secrets.Port)
	logger.Printf("[*] %s listening on address: %s", config.ServiceName, address)

	server.ExecGRPCServer(address, coreService, logger)
}
