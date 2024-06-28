package api

// Registers routes and their handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"main/config"
	"main/repository"
	"main/services"
	"net/http"
)

type Dependencies struct {
	PlayerRepository      *repository.Player
	ChallengeRepository   *repository.Challenger
	TransactionRepository *repository.Transaction

	RegistrationHandler *RegistrationHandler
	LoginHandler        *LoginHandler
	PlayersHandler      *PlayersHandler
	ChallengeHandler    *ChallengeHandler
	TransactionHandler  *TransactionHandler
}

var dependencies *Dependencies

func LoadServerDependencies(deps *Dependencies) {
	dependencies = deps
}

func StartServer() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello to rock, paper, scissors.",
		})
	})

	authorized := router.Group("/")
	authorized.Use(services.AuthenticateUser)

	// Register new players
	router.POST("/registration", dependencies.RegistrationHandler.Handle)
	// Try to log in a player
	router.POST("/login", dependencies.LoginHandler.Handle)
	// Find available players
	authorized.GET("/players", dependencies.PlayersHandler.GetAllPlayers)
	// Deposit or withdraw
	authorized.POST("/funds", dependencies.PlayersHandler.TransferFunds)
	// Challenger player
	authorized.POST("/challenge", dependencies.ChallengeHandler.Create)
	// Settle challenge
	authorized.POST("/challenge/settle", dependencies.ChallengeHandler.Settle)
	// Decline challenge
	authorized.POST("/challenge/decline", dependencies.ChallengeHandler.Decline)
	// Get pending challenges
	authorized.GET("/challenge/pending", dependencies.ChallengeHandler.GetPendingChallenges)
	// Get pending transactions
	authorized.GET("/transactions", dependencies.TransactionHandler.GetTransactionsByUsername)

	err := router.Run(fmt.Sprintf(":%s", config.Settings.ServerPort))
	if err != nil {
		panic(err.Error())
	}
}
