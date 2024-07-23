package sqlite

import (
	"context"
	"database/sql"
	"gitbeam.commit.monitor/models"
	"gitbeam.commit.monitor/repository"
	_ "github.com/mattn/go-sqlite3"
)

const cronTrackerTableSetup = `
CREATE TABLE IF NOT EXISTS cron_tasks (
		repo_name TEXT,
		owner_name TEXT,
		from_date DATETIME,
		to_date DATETIME,
		duration_in_hours INTEGER,
		UNIQUE (repo_name, owner_name)
)
`

func scanCronTrackerRow(row *sql.Row) (*models.MonitorRepositoryCommitConfig, error) {
	var cronTracker models.MonitorRepositoryCommitConfig
	var err error
	if err = row.Scan(
		&cronTracker.RepoName,
		&cronTracker.OwnerName,
		&cronTracker.FromDate,
		&cronTracker.ToDate,
		&cronTracker.DurationInHours,
	); err != nil {
		return nil, err
	}

	return &cronTracker, nil
}

func scanCronTrackerRows(rows *sql.Rows) (*models.MonitorRepositoryCommitConfig, error) {
	var cronTracker models.MonitorRepositoryCommitConfig
	var err error
	if err = rows.Scan(
		&cronTracker.RepoName,
		&cronTracker.OwnerName,
		&cronTracker.FromDate,
		&cronTracker.ToDate,
		&cronTracker.DurationInHours,
	); err != nil {
		return nil, err
	}

	return &cronTracker, nil
}

func (s sqliteRepo) ListMonitorConfig(ctx context.Context) ([]*models.MonitorRepositoryCommitConfig, error) {
	querySQL := `SELECT * FROM cron_tasks`

	rows, err := s.dataStore.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, err
	}

	var list []*models.MonitorRepositoryCommitConfig
	defer rows.Close()
	for rows.Next() {
		item, err := scanCronTrackerRows(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (s sqliteRepo) SaveMonitorConfigs(ctx context.Context, payload models.MonitorRepositoryCommitConfig) error {
	insertSQL := `
        INSERT INTO cron_tasks (
			repo_name,
			owner_name,
			from_date,
			to_date,
			duration_in_hours,
		)
        VALUES (?, ?, ?, ?)`

	_, err := s.dataStore.ExecContext(ctx, insertSQL,
		payload.RepoName,
		payload.OwnerName,
		payload.FromDate,
		payload.ToDate,
		payload.DurationInHours,
	)
	return err
}

func (s sqliteRepo) GetMonitorConfig(ctx context.Context, owner models.OwnerAndRepoName) (*models.MonitorRepositoryCommitConfig, error) {
	row := s.dataStore.QueryRowContext(ctx,
		`SELECT * from cron_tasks WHERE owner_name = ? AND repo_name = ? LIMIT 1`, owner.OwnerName, owner.RepoName)
	return scanCronTrackerRow(row)
}

func (s sqliteRepo) DeleteMonitorConfig(ctx context.Context, owner models.OwnerAndRepoName) error {
	_, err := s.dataStore.ExecContext(ctx,
		`DELETE from cron_tasks WHERE owner_name = ? AND repo_name = ?`, owner.OwnerName, owner.RepoName)

	return err
}

func NewSqliteCronStore(dbName string) (repository.CronServiceStore, error) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(cronTrackerTableSetup); err != nil {
		return nil, err
	}
	return &sqliteRepo{
		dataStore: db,
	}, nil
}
