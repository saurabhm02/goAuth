package main

import (
	"log"
	"net/http"
	"os"

	"goAuth/internals/config"
	"goAuth/internals/handler"
	"goAuth/internals/repository"
	"goAuth/internals/routes"
	"goAuth/internals/types"
)

func main() {
	cfg, err := config.Load(configPath())
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	store := buildProjectStore(cfg)

	router := routes.NewRouter(
		handler.NewAuthHandler(),
		store,
		routes.AuthMiddlewareFromProject(),
	)

	port := os.Getenv(types.EnvPort)
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	srv := &http.Server{Addr: ":" + port, Handler: router}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func configPath() string {
	if p := os.Getenv(types.EnvConfigPath); p != "" {
		return p
	}
	return "config.yaml"
}

func buildProjectStore(cfg *config.Config) map[string]interface{} {
	store := make(map[string]interface{})

	for projectID, p := range cfg.Projects {
		proj := new(config.ProjectConfig)
		*proj = p

		repo, err := repository.NewPostgresUserRepository(proj)
		if err != nil {
			log.Fatalf("project %s: %v", projectID, err)
		}

		store[projectID] = &types.ProjectContext{Repo: repo}
	}

	return store
}
