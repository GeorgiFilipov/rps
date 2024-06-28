package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // PostgreSQL driver
	"main/api"
	"main/config"
	"main/repository"
)

func main() {
	config.LoadConfig()
	db := createDBConnection(config.Settings)
	defer db.Close()

	// Inject dependencies
	var dependencies api.Dependencies
	dependencies.PlayerRepository = repository.NewPlayerRepository(db)
	dependencies.ChallengeRepository = repository.NewChallengeRepository(db)
	dependencies.TransactionRepository = repository.NewTransactionRepository(db)

	dependencies.RegistrationHandler = api.NewRegistrationHandler(dependencies.PlayerRepository, dependencies.TransactionRepository)
	dependencies.LoginHandler = api.NewLoginHandler(dependencies.PlayerRepository)
	dependencies.PlayersHandler = api.NewFindPlayersHandler(dependencies.PlayerRepository, dependencies.TransactionRepository)
	dependencies.ChallengeHandler = api.NewChallengeHandler(dependencies.ChallengeRepository, dependencies.PlayerRepository, dependencies.TransactionRepository)
	dependencies.TransactionHandler = api.NewTransactionHandler(dependencies.TransactionRepository)

	api.LoadServerDependencies(&dependencies)

	api.StartServer()
}

func createDBConnection(config config.Config) *sql.DB {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=localhost port=5432 sslmode=disable",
		config.DBUser, config.DBPass, config.DBName)

	fmt.Println(connStr)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	_, err = db.Exec("SET search_path TO public")
	if err != nil {
		db.Close()
		panic(fmt.Errorf("failed to set search_path: %v", err))
	}

	return db
}
