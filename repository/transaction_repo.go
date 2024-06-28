package repository

import (
	"database/sql"
	"github.com/sirupsen/logrus"
	"log"
	"main/model"
	"time"
)

type Transaction struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *Transaction {
	return &Transaction{
		db: db,
	}
}

func (repository *Transaction) AddTransaction(amount int, reason, user string) error {
	query := `
		INSERT INTO transaction (amount, reason, username)
		VALUES ($1, $2, $3)
		RETURNING id, timestamp
	`

	var id int
	var timestamp time.Time
	err := repository.db.QueryRow(query, amount, reason, user).Scan(&id, &timestamp)
	if err != nil {
		logrus.Errorf("Error inserting transaction: %v", err)
		return err
	}

	logrus.Printf("Inserted transaction with ID %d and timestamp %v", id, timestamp)
	return nil
}

func (repository *Transaction) GetTransactionsByUsername(username string) ([]model.Transaction, error) {
	query := `
        SELECT username, amount, reason, timestamp
        FROM transaction
        WHERE username = $1
    `

	rows, err := repository.db.Query(query, username)
	if err != nil {
		log.Printf("Error fetching transactions: %v", err)
		return nil, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var transaction model.Transaction
		if err = rows.Scan(
			&transaction.Username,
			&transaction.Amount,
			&transaction.Reason,
			&transaction.Timestamp,
		); err != nil {
			log.Printf("Error scanning transaction: %v", err)
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		return nil, err
	}

	return transactions, nil
}
