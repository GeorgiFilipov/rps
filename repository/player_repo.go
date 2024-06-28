package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"main/internal"
	"main/model"
)

type Player struct {
	db *sql.DB
}

func NewPlayerRepository(db *sql.DB) *Player {
	return &Player{
		db: db,
	}
}

// RegisterPlayer tries to register the player if they don't already exist
func (repository *Player) RegisterPlayer(playerRegistration *model.PlayerRegistrationRequest) (*model.Player, error) {
	err := repository.validatePlayerRegistration(playerRegistration)
	if err != nil {
		logrus.Error("Failed to validate player registration")
		return nil, err
	}

	salt, err := internal.GenerateRandomSalt()
	if err != nil {
		logrus.Errorf("Failed to generate salt: %s", err)
		return nil, err
	}

	hashed, err := internal.HashPassword(playerRegistration.Password, salt)
	if err != nil {
		logrus.Errorf("Failed to hash password: %s", err)
		return nil, err
	}

	newPlayer := &model.Player{
		Username: playerRegistration.Username,
		Password: hashed,
		Salt:     salt,
		Balance:  playerRegistration.Deposit,
	}

	_, err = repository.db.Exec(
		"INSERT INTO player (username, password, salt, balance) VALUES ($1, $2, $3, $4)",
		newPlayer.Username, newPlayer.Password, newPlayer.Salt, newPlayer.Balance,
	)
	if err != nil {
		logrus.Errorf("Failed to register player: %s", err)
		return nil, err
	}

	return newPlayer, nil
}

func (repository *Player) FindPlayerWithDetails(username string) (*model.Player, error) {
	var player model.Player
	err := repository.db.QueryRow(
		"SELECT username, password, salt, balance FROM player WHERE username = $1",
		username,
	).Scan(&player.Username, &player.Password, &player.Salt, &player.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logrus.Infof("Player not found: %s", username)
			return nil, nil
		}
		logrus.Errorf("Failed to find player: %s", err)
		return nil, err
	}
	return &player, nil
}

func (repository *Player) Exists(username string) (bool, error) {
	var exists bool
	err := repository.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM player WHERE username = $1)",
		username,
	).Scan(&exists)
	if err != nil {
		logrus.Errorf("Failed to check if player exists: %s", err)
		return false, err
	}
	return exists, nil
}

func (repository *Player) GetPlayerBalance(username string) (int, error) {
	player, err := repository.FindPlayerWithDetails(username)
	if err != nil {
		return 0, err
	}
	return player.Balance, nil
}

// AddPlayerBalance adds balance to the player's current balance.
func (repository *Player) AddPlayerBalance(username string, amountToAdd int) error {

	// Retrieve current balance
	var currentBalance int
	err := repository.db.QueryRow("SELECT balance FROM player WHERE username = $1", username).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to fetch current balance: %v", err)
	}

	// Calculate new balance
	newBalance := currentBalance + amountToAdd
	if newBalance < 0 {
		return errors.New("insufficient balance")
	}

	// Update player balance
	_, err = repository.db.Exec("UPDATE player SET balance = $1 WHERE username = $2", newBalance, username)
	if err != nil {
		return fmt.Errorf("failed to update balance: %v", err)
	}

	return nil
}

// SubtractPlayerBalance subtracts balance from the player's current balance.
func (repository *Player) SubtractPlayerBalance(username string, amountToSubtract int) error {
	// Retrieve current balance
	var currentBalance int
	err := repository.db.QueryRow("SELECT balance FROM player WHERE username = $1", username).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to fetch current balance: %v", err)
	}

	// Calculate new balance
	newBalance := currentBalance - amountToSubtract
	if newBalance < 0 {
		return errors.New("insufficient balance")
	}

	// Update player balance
	_, err = repository.db.Exec("UPDATE player SET balance = $1 WHERE username = $2", newBalance, username)
	if err != nil {
		return fmt.Errorf("failed to update balance: %v", err)
	}

	return nil
}

func (repository *Player) GetAllPlayerUsernames() ([]string, error) {
	query := `
        SELECT username
        FROM player
    `

	rows, err := repository.db.Query(query)
	if err != nil {
		log.Printf("Error fetching usernames: %v", err)
		return nil, err
	}
	defer rows.Close()

	var usernames []string
	for rows.Next() {
		var username string
		if err = rows.Scan(&username); err != nil {
			logrus.Errorf("Error scanning username: %v", err)
			return nil, err
		}
		usernames = append(usernames, username)
	}

	if err = rows.Err(); err != nil {
		logrus.Errorf("Error iterating over rows: %v", err)
		return nil, err
	}

	return usernames, nil
}

func (repository *Player) validatePlayerRegistration(playerRegistration *model.PlayerRegistrationRequest) error {
	err := internal.ValidatePlayerUsername(playerRegistration.Username)
	if err != nil {
		return err
	}

	err = internal.ValidatePlayerPassword(playerRegistration.Password)
	if err != nil {
		return err
	}

	err = internal.ValidateMinimumPlayerDeposit(playerRegistration.Deposit)
	if err != nil {
		return err
	}

	exists, err := repository.Exists(playerRegistration.Username)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("player already exists")
	}

	return nil
}
