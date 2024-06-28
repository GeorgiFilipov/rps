package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"main/config"
	"main/model"
	"main/repository"
	"net/http"
)

type RegistrationHandler struct {
	players      *repository.Player
	transactions *repository.Transaction
}

func NewRegistrationHandler(playerRepository *repository.Player,
	transactionRepository *repository.Transaction) *RegistrationHandler {
	return &RegistrationHandler{
		players:      playerRepository,
		transactions: transactionRepository,
	}
}

// Handle Registers users that do not have overlapping email addresses
func (regHandler *RegistrationHandler) Handle(context *gin.Context) {
	var registration model.PlayerRegistrationRequest

	err := context.BindJSON(&registration)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if registration.Deposit < config.Settings.MinimumDeposit {
		logrus.Error("Deposit below minimum")
		context.AbortWithStatusJSON(http.StatusBadRequest,
			fmt.Sprintf("Deposit must be at least %d", config.Settings.MinimumDeposit))
		return
	}

	player, err := regHandler.players.RegisterPlayer(&registration)
	if err != nil {
		logrus.Errorf("Unable to register player with username: %s , username already exists", registration.Username)
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	logrus.Infof("Registered player with username %s", registration.Username)

	err = regHandler.transactions.AddTransaction(registration.Deposit, model.ReasonDeposit, registration.Username)
	if err != nil {
		logrus.Error("Failed to log transaction for deposit")
	}

	context.JSON(http.StatusCreated, gin.H{"Message": fmt.Sprintf("Player with username: %s registered.", player.Username)})
}
