package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"go/build"
	"os"
	"path/filepath"
)

const ServiceName = "gitbeam.commit.monitor"

type Secrets struct {
	CommitDatabaseName string `json:"COMMIT_DATABASE_NAME"`
	CronDatabaseName   string `json:"CRON_DATABASE_NAME"`
	Port               string
}

var ss Secrets

func init() {
	importPath := fmt.Sprintf("%s/config", ServiceName)
	p, err := build.Default.Import(importPath, "", build.FindOnly)
	if err == nil {
		env := filepath.Join(p.Dir, "../.env")
		_ = godotenv.Load(env)
	}

	ss = Secrets{}
	ss.CommitDatabaseName = os.Getenv("COMMIT_DATABASE_NAME")
	ss.CronDatabaseName = os.Getenv("CRON_DATABASE_NAME")
	if ss.Port = os.Getenv("PORT"); ss.Port == "" {
		ss.Port = "80"
	}
}

// GetSecrets is used to get value from the Secrets runtime.
func GetSecrets() Secrets {
	return ss
}
