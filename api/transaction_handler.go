package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"main/repository"
	"main/services"
	"net/http"
)

type TransactionHandler struct {
	transactions *repository.Transaction
}

func NewTransactionHandler(transactions *repository.Transaction) *TransactionHandler {
	return &TransactionHandler{
		transactions: transactions,
	}
}

func (transactionHandler *TransactionHandler) GetTransactionsByUsername(context *gin.Context) {
	userName := services.GetSubjectFromContext(context)
	userTransactions, err := transactionHandler.transactions.GetTransactionsByUsername(userName)
	if err != nil {
		logrus.Error("Unable to get transactions for user")
		context.AbortWithStatusJSON(http.StatusInternalServerError, "Failed to retrieve transactions")
		return
	}

	context.JSON(http.StatusOK, userTransactions)
}
