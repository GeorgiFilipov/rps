package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"main/config"
	"main/model"
	"main/repository"
	"main/services"
	"net/http"
)

type ChallengeHandler struct {
	challenges   *repository.Challenger
	players      *repository.Player
	transactions *repository.Transaction
}

func NewChallengeHandler(challengeRepository *repository.Challenger,
	playerRepository *repository.Player,
	transactions *repository.Transaction) *ChallengeHandler {
	return &ChallengeHandler{
		challenges:   challengeRepository,
		players:      playerRepository,
		transactions: transactions,
	}
}

func (challengeHandler *ChallengeHandler) Create(context *gin.Context) {

	var challengeRequest model.ChallengeRequest
	err := context.BindJSON(&challengeRequest)
	if err != nil {
		logrus.Error("Unable to bind challenge request body")
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	// Make sure that the challenger and the player are available
	challenger := services.GetSubjectFromContext(context)
	if challenger == "" {
		context.AbortWithStatusJSON(http.StatusBadRequest, "No user in token")
		return
	}

	if isValidChoice(challengeRequest.Choice) {
		logrus.Error("Invalid choice")
		context.AbortWithStatusJSON(http.StatusBadRequest, "Invalid choice")
		return
	}

	exists, err := challengeHandler.players.Exists(challengeRequest.Opponent)
	if !exists || err != nil {
		context.AbortWithStatusJSON(http.StatusNotFound, "Opponent does not exist")
		return
	}

	// Check if enough balance is available
	balance, err := challengeHandler.players.GetPlayerBalance(challenger)
	if err != nil {
		logrus.Error("Unable to get player balance err")
		context.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if challengeRequest.Bet < config.Settings.MinimumBet {
		logrus.Error("Bet too low")
		context.AbortWithStatusJSON(http.StatusBadRequest, "Bet amount is too low")
		return
	}

	if balance < challengeRequest.Bet {
		logrus.Error("Attempting to bet with too low balance")
		context.AbortWithStatusJSON(http.StatusBadRequest, "Not enough balance to place bet")
		return
	}

	err = challengeHandler.players.SubtractPlayerBalance(challenger, challengeRequest.Bet)
	if err != nil {
		logrus.Error("Failed to subtract balance")
		context.AbortWithStatusJSON(http.StatusInternalServerError, "Unable to take funds")
		return
	}

	_ = challengeHandler.transactions.AddTransaction(-challengeRequest.Bet, model.ReasonBet, challenger)

	challengeId, err := challengeHandler.challenges.CreateChallenge(
		challenger, challengeRequest.Opponent, challengeRequest.Choice, challengeRequest.Bet)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"ChallengeId": challengeId,
	})

}

func (challengeHandler *ChallengeHandler) Settle(context *gin.Context) {
	var challengeSettleRequest model.ChallengeSettleRequest
	err := context.BindJSON(&challengeSettleRequest)
	if err != nil {
		logrus.Error("Unable to bind challenge request body")
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	userName := services.GetSubjectFromContext(context)

	// Find challenge
	challenge, err := challengeHandler.challenges.GetChallengeByID(challengeSettleRequest.ChallengeId)
	if err != nil {
		logrus.Error("Unable to get challenge err: %s", err.Error())
		context.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Check if challenge belongs to the player trying to resolve it
	if challenge.Opponent != userName {
		logrus.Error("Attempting to resolve challenge that belongs to another player, aborting")
		context.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "not allowed to settle challenge"})
		return
	}

	if challenge.State != "pending" {
		logrus.Error("Attempting to settle already settled challenge")
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "challenge already settled"})
		return
	}

	// Validate choice
	if isValidChoice(challengeSettleRequest.Choice) {
		logrus.Error("Invalid choice")
		context.AbortWithStatusJSON(http.StatusBadRequest, "Invalid choice")
		return
	}

	// Get balance
	currentBalance, err := challengeHandler.players.GetPlayerBalance(userName)
	if err != nil {
		logrus.Error("Unable to get player balance")
		context.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Make sure there's enough balance
	if currentBalance < challenge.Bet {
		logrus.Errorf("Unable to accept challenge, not enough funds")
		context.AbortWithStatusJSON(http.StatusBadRequest, "Not enough funds")
		return
	}

	// Subtract funds before proceeding
	err = challengeHandler.players.SubtractPlayerBalance(userName, challenge.Bet)
	if err != nil {
		logrus.Errorf("Unable to update player balance")
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unable to settle challenge, try again"})
		return
	}

	_ = challengeHandler.transactions.AddTransaction(-challenge.Bet, model.ReasonBet, userName)

	// Opponent's choice
	challengerChoice := challenge.Choice
	// Current player's choice
	opponentChoice := challengeSettleRequest.Choice

	winner := determineWinner(challengerChoice, opponentChoice)

	// Restore the subtracted money to the challenger
	challengeWinner := ""
	message := ""
	if winner == "draw" {
		err = challengeHandler.players.AddPlayerBalance(challenge.Challenger, challenge.Bet)
		_ = challengeHandler.transactions.AddTransaction(challenge.Bet, model.ReasonRefund, challenge.Challenger)
		message = fmt.Sprintf("Draw both players picked :%s ", model.ChoiceToString(opponentChoice))
	} else if winner == "opponent" {
		err = challengeHandler.players.AddPlayerBalance(userName, challenge.Bet)
		challengeWinner = userName
		_ = challengeHandler.transactions.AddTransaction(challenge.Bet, model.ReasonWin, challengeWinner)
		message = fmt.Sprintf("Winner :%s with %s against %s", challengeWinner, model.ChoiceToString(challengerChoice), model.ChoiceToString(opponentChoice))
	} else if winner == "challenger" {
		// Gets his initial deposit and his opponent's money
		err = challengeHandler.players.AddPlayerBalance(challenge.Challenger, challenge.Bet*2)
		challengeWinner = challenge.Challenger
		_ = challengeHandler.transactions.AddTransaction(challenge.Bet*2, model.ReasonWin, challengeWinner)
		message = fmt.Sprintf("Winner :%s with %s against %s", challengeWinner, model.ChoiceToString(challengerChoice), model.ChoiceToString(opponentChoice))
	}

	if err != nil {
		logrus.Errorf("Unable to update player balance")
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unable to settle challenge, try again"})
		return
	}

	err = challengeHandler.challenges.UpdateChallenge(model.ChallengeSettled, challengeWinner, challenge.ChallengeId)
	if err != nil {
		logrus.Errorf("Unable to update challenge: %s", challenge.ChallengeId)
		context.AbortWithStatusJSON(http.StatusInternalServerError, "Failed to update challenge")
		return
	}

	context.JSON(http.StatusOK, model.ChallengeResponse{
		Winner:    winner,
		WinAmount: challenge.Bet,
		Message:   message,
	})
}

func (challengeHandler *ChallengeHandler) Decline(context *gin.Context) {
	var challengeDeclineRequest model.ChallengeDeclineRequest
	err := context.BindJSON(&challengeDeclineRequest)
	if err != nil {
		logrus.Error("Unable to bind challenge request body")
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	userName := services.GetSubjectFromContext(context)

	// Find challenge
	challenge, err := challengeHandler.challenges.GetChallengeByID(challengeDeclineRequest.ChallengeId)
	if err != nil {
		logrus.Error("Unable to get challenge err: %s", err.Error())
		context.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	// Check if challenge was initiated by one of the two players
	if challenge.Opponent != userName || challenge.Challenger != userName {
		logrus.Error("Challenger does not belong to player")
		context.AbortWithStatusJSON(http.StatusForbidden, "Challenger does not belong to player")
		return
	}

	// Can only decline pending challenges
	if challenge.State != model.ChallengePending {
		logrus.Error("Challenger is already settled")
		context.AbortWithStatusJSON(http.StatusForbidden, "Challenger is already settled")
		return
	}

	err = challengeHandler.challenges.UpdateChallenge(model.ChallengeDeclined, "", challenge.ChallengeId)
	if err != nil {
		logrus.Error("Failed to decline challenge, refund challenger manually")
		context.AbortWithStatusJSON(http.StatusInternalServerError, "Failed to decline challenge, try again")
		return
	}

	// Try to return funds to the original challenger
	err = challengeHandler.players.AddPlayerBalance(challenge.Challenger, challenge.Bet)
	if err != nil {
		logrus.Error("Failed to refund challenger, please update manually")
		context.AbortWithStatusJSON(http.StatusInternalServerError, "Failed to refund challenger")
		return
	}

	context.JSON(http.StatusOK, "Successfully declined challenge")

}

// GetPendingChallenges Retrieves the pending challenges for a user
func (challengeHandler *ChallengeHandler) GetPendingChallenges(context *gin.Context) {
	userName := services.GetSubjectFromContext(context)
	pendingChallenges, err := challengeHandler.challenges.GetPendingChallenges(userName)
	if err != nil {
		logrus.Error("Failed to retrieve pending challenges")
		context.AbortWithStatusJSON(http.StatusInternalServerError, "Failed to retrieve pending challenges")
		return
	}

	context.JSON(http.StatusOK, pendingChallenges)

}

// Only the opponent resolves the challenge
func determineWinner(challengerChoice int, opponentChoice int) string {
	if challengerChoice == opponentChoice {
		return "draw"
	}

	switch challengerChoice {
	case model.ChoiceRock:
		if opponentChoice == model.ChoiceScissors {
			return "challenger"
		} else {
			return "opponent"
		}
	case model.ChoicePaper:
		if opponentChoice == model.ChoiceRock {
			return "challenger"
		} else {
			return "opponent"
		}
	case model.ChoiceScissors:
		if opponentChoice == model.ChoicePaper {
			return "challenger"
		} else {
			return "opponent"
		}
	}

	return ""
}

func isValidChoice(choice int) bool {
	return choice < model.ChoiceRock || choice > model.ChoiceScissors
}
