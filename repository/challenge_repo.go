package repository

import (
	"database/sql"
	"github.com/sirupsen/logrus"
	"main/model"
	"time"
)

type Challenger struct {
	db *sql.DB
}

func NewChallengeRepository(db *sql.DB) *Challenger {
	return &Challenger{db: db}
}

// CreateChallenge inserts a new challenge into the database and returns its id
func (repository *Challenger) CreateChallenge(challenger string, opponent string, choice int, bet int) (int, error) {
	query := `
        INSERT INTO challenge (challenger, opponent, choice, bet, state)
        VALUES ($1, $2, $3, $4, $5) RETURNING challenge_id
    `

	var challengeId int

	err := repository.db.QueryRow(query, challenger, opponent, choice, bet, model.ChallengePending).Scan(&challengeId)
	if err != nil {
		logrus.Errorf("Error inserting challenge: %v", err)
		return 0, err
	}

	return challengeId, nil
}

// GetChallengeByID retrieves a challenge by its ID
func (repository *Challenger) GetChallengeByID(challengeID string) (*model.Challenge, error) {
	query := `
        SELECT challenge_id, challenger, opponent, choice, bet, state, time_created, time_settled
        FROM challenge
        WHERE challenge_id = $1
    `

	var challenge model.Challenge
	var timeSettled sql.NullTime
	err := repository.db.QueryRow(query, challengeID).Scan(
		&challenge.ChallengeId,
		&challenge.Challenger,
		&challenge.Opponent,
		&challenge.Choice,
		&challenge.Bet,
		&challenge.State,
		&challenge.TimeCreated,
		&timeSettled,
	)

	if err != nil {
		logrus.Errorf("Error fetching challenge: %v", err)
		return nil, err
	}

	if timeSettled.Valid {
		challenge.TimeSettled = timeSettled.Time
	} else {
		challenge.TimeSettled = time.Time{}
	}

	return &challenge, nil
}

// GetPendingChallenges retrieves all pending challenges where the user is listed as an opponent
func (repository *Challenger) GetPendingChallenges(username string) ([]model.PendingChallenge, error) {
	query := `
        SELECT challenge_id, challenger, bet, time_created
        FROM challenge
        WHERE opponent = $1 AND challenge.state = 'pending'
    `

	rows, err := repository.db.Query(query, username)
	if err != nil {
		logrus.Errorf("Error fetching challenges: %v", err)
		return nil, err
	}
	defer rows.Close()

	var challenges []model.PendingChallenge
	for rows.Next() {
		var challenge model.PendingChallenge
		err = rows.Scan(
			&challenge.ChallengeId,
			&challenge.Challenger,
			&challenge.Bet,
			&challenge.TimeCreated,
		)
		if err != nil {
			logrus.Errorf("Error scanning challenge: %v", err)
			return nil, err
		}
		challenges = append(challenges, challenge)
	}

	if err = rows.Err(); err != nil {
		logrus.Errorf("Error with rows: %v", err)
		return nil, err
	}

	return challenges, nil
}

// UpdateChallenge updates the status and time_settled of a challenge
func (repository *Challenger) UpdateChallenge(state string, winner string, challengeId string) error {
	query := `
        UPDATE challenge
        SET state = $1, time_settled = $2, winner = $3
        WHERE challenge_id = $4
    `

	_, err := repository.db.Exec(query, state, time.Now(), winner, challengeId)
	if err != nil {
		logrus.Errorf("Error updating challenge: %v", err)
		return err
	}

	return nil
}
