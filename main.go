package main

import (
	"flag"
	"os"
	"time"

	"github.com/martenwallewein/todo-service/api"
	"github.com/martenwallewein/todo-service/pkg/git"
	log "github.com/sirupsen/logrus"
)

var (
	laddr           = flag.String("addr", ":8880", "Local address for the HTTP API")
	loglevel        = flag.String("loglevel", "TRACE", "Log-level (ERROR|WARN|INFO|DEBUG|TRACE)")
	initialSeedFile = flag.String("initialSeedFile", "", "Run one-time seeds passing path to a valid JSON seed file")
)

func configureLogging() error {
	l, err := log.ParseLevel(*loglevel)
	if err != nil {
		return err
	}
	log.SetLevel(l)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.RFC3339Nano,
		ForceColors:     true,
	})
	return nil
}

func main() {
	flag.Parse()
	if err := configureLogging(); err != nil {
		log.Fatal(err)
	}
	/*
		err := dbs.InitializeDatabaseLayer()
		if err != nil {
			log.Fatal(err)
		}

		if initialSeedFile != nil && *initialSeedFile != "" {
			if err = seeds.RunSeeds(*initialSeedFile); err != nil {
				log.Fatal(err)
			}
		}
	*/

	path := os.Getenv("TODO_REPO_PATH")
	if path == "" {
		log.Fatal("Missing repo path")
	}

	repoUrl := os.Getenv("TODO_REPO_GIT_URL")
	if path == "" {
		log.Fatal("Missing repo path")
	}

	// Repo does not exist yet
	if _, err := os.Stat(path); err != nil {

		err := os.MkdirAll(path, 0775)
		if err != nil {
			log.Fatal("Failed to create base repo dir ", err)
		}
		_, err = git.Clone(repoUrl, path)
		if err != nil {
			log.Fatal(err)
		}
	}

	api := api.NewRESTApiV1(path)
	if err := api.Serve(*laddr); err != nil {
		log.Fatal(err)
	}
}
