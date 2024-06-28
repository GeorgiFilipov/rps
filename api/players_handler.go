package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"main/model"
	"main/repository"
	"main/services"
	"net/http"
)

type PlayersHandler struct {
	players      *repository.Player
	transactions *repository.Transaction
}

func NewFindPlayersHandler(players *repository.Player, transactions *repository.Transaction) *PlayersHandler {
	return &PlayersHandler{players: players, transactions: transactions}
}

func (playersHandler *PlayersHandler) GetAllPlayers(context *gin.Context) {
	usernames, err := playersHandler.players.GetAllPlayerUsernames()
	if err != nil {
		logrus.Errorf("Unable to get players err: %s", err.Error())
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	context.JSON(http.StatusFound, usernames)
}

func (playersHandler *PlayersHandler) TransferFunds(context *gin.Context) {

	var transactionRequest model.TransactionRequest
	err := context.BindJSON(&transactionRequest)
	if err != nil {
		logrus.Errorf("Unable to bind %v", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	userName := services.GetSubjectFromContext(context)

	switch transactionRequest.Reason {
	case model.ReasonDeposit:
		err = playersHandler.players.SubtractPlayerBalance(userName, transactionRequest.Amount)
	case model.ReasonWithdrawal:
		err = playersHandler.players.SubtractPlayerBalance(userName, transactionRequest.Amount)
	default:
		logrus.Error("Wrong reason for funds transfer")
		context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong reason for funds transfer")
		return
	}

	if err != nil {
		logrus.Error("Unable to transfer funds")
		context.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"message": "Unable to transfer funds", "error": err.Error()})
		return
	}

	if transactionRequest.Reason == model.ReasonDeposit {
		_ = playersHandler.transactions.AddTransaction(transactionRequest.Amount, model.ReasonDeposit, userName)
	}
	_ = playersHandler.transactions.AddTransaction(-transactionRequest.Amount, model.ReasonWithdrawal, userName)

	context.JSON(http.StatusOK, fmt.Sprintf("Transaction %s is successful", transactionRequest.Reason))
	return

}
