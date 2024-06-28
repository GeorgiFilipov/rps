package model

import (
	"time"
)

const (
	ReasonDeposit    = "deposit"
	ReasonWithdrawal = "withdrawal"
	ReasonWin        = "win"
	ReasonRefund     = "refund"
	ReasonBet        = "bet"
)

type Transaction struct {
	ID        int       `json:"id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Amount    int       `json:"amount"`
	Reason    string    `json:"reason"`
	Username  string    `json:"username"`
}

type TransactionRequest struct {
	Reason string `json:"reason" binding:"required"`
	Amount int    `json:"amount" binding:"required"`
}
