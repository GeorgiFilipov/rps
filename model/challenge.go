package model

import "time"

const (
	ChoiceRock     = 1
	ChoicePaper    = 2
	ChoiceScissors = 3
)

const (
	ChallengePending  = "pending"
	ChallengeSettled  = "settled"
	ChallengeDeclined = "declined"
)

// ChallengeRequest creates a challenge request
type ChallengeRequest struct {
	Opponent string `json:"opponent" binding:"required"`
	Choice   int    `json:"choice" binding:"required"`
	Bet      int    `json:"bet" binding:"required"`
}

// Challenge takes a challenge request and adds it to the pending challenges
type Challenge struct {
	ChallengeId string `json:"challenge_id" binding:"required"`
	Challenger  string `json:"challenger" binding:"required"`
	ChallengeRequest
	State       string    `json:"state" binding:"required"`
	TimeCreated time.Time `json:"time_created" binding:"required"`
	TimeSettled time.Time `json:"time_settled"`
	Winner      string    `json:"winner"`
}

type PendingChallenge struct {
	ChallengeId string    `json:"challenge_id"`
	Challenger  string    `json:"challenger" `
	Bet         int       `json:"bet"`
	TimeCreated time.Time `json:"time_created"`
}

type ChallengeDeclineRequest struct {
	ChallengeId string `json:"challenge_id" binding:"required"`
}

type ChallengeSettleRequest struct {
	ChallengeId string `json:"challenge_id" binding:"required"`
	Choice      int    `json:"bet_choice"`
}

type ChallengeResponse struct {
	Winner    string `json:"winner"`
	WinAmount int    `json:"winAmount"`
	Message   string `json:"message"`
}

func ChoiceToString(choice int) string {
	switch choice {
	case 1:
		return "rock"
	case 2:
		return "paper"
	case 3:
		return "scissors"
	default:
		return ""

	}
}
