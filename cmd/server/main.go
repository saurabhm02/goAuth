package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"goAuth/internals/config"
	"goAuth/internals/handler"
	"goAuth/internals/repository"
	"goAuth/internals/types"
	"goAuth/internals/worker"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load(configPath())
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	store := buildProjectStore(cfg)

	router := handler.NewRouter(
		handler.NewAuthHandler(),
		store,
		handler.AuthMiddlewareFromProject(),
	)

	go worker.StartOTPCleanup(store)

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

func openDB(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	return db
}

func buildProjectStore(cfg *config.Config) map[string]interface{} {
	store := make(map[string]interface{})

	for projectID, p := range cfg.Projects {
		proj := new(config.ProjectConfig)
		*proj = p
		db := openDB(proj.Database.DSN)

		repo := repository.NewPostgresUserRepository(db, proj.Database.UserTable)
		pc := &types.ProjectContext{Repo: repo, OTP: p.OTP}

		if p.OTP {
			otpRepo := repository.NewPostgresOTPRepository(db, proj.Database.OtpTable)
			pc.OTPRepo = otpRepo
		}
		store[projectID] = pc
	}

	return store
}
