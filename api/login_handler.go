package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"main/internal"
	"main/model"
	"main/repository"
	"main/services"
	"net/http"
)

type LoginHandler struct {
	playerRepository *repository.Player
}

func NewLoginHandler(playerRepository *repository.Player) *LoginHandler {
	return &LoginHandler{playerRepository: playerRepository}
}

// Handle attempts to log users in with username and password
func (loginHandler *LoginHandler) Handle(context *gin.Context) {
	var loginData model.PlayerLoginRequest
	err := context.BindJSON(&loginData)
	if err != nil {
		logrus.Errorf("Unable to bind JSON err: %s", err.Error())
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	err, statusCode := loginHandler.checkLoginData(loginData)
	if err != nil {
		context.AbortWithStatusJSON(statusCode, err.Error())
		return
	}

	// Create JWT for the user and return it
	token, err := services.GenerateJWT(loginData.Username)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, errors.New("unable to create token"))
		return
	}
	logrus.Infof("Created a new token for username: %s", loginData.Username)

	context.JSON(http.StatusCreated, gin.H{
		"token": token,
	})

}

// Check if token is valid and not expired
// if expired, check credentials
// if credentials are correct, create a new token and add it to the token repo
// else throw error
func (loginHandler *LoginHandler) checkLoginData(loginData model.PlayerLoginRequest) (error, int) {
	err := internal.ValidatePlayerUsername(loginData.Username)
	if err != nil {
		logrus.Warning("Logging in user was not successful, validation failed")
		return err, http.StatusBadRequest
	}

	exists, err := loginHandler.playerRepository.Exists(loginData.Username)
	if err != nil {
		logrus.Error(err)
		return err, http.StatusInternalServerError
	}
	if !exists {
		logrus.Errorf("Unable to find player by username: %s", loginData.Username)
		return errors.New("username or password mismatch"), http.StatusUnauthorized
	}

	playerDetails, err := loginHandler.playerRepository.FindPlayerWithDetails(loginData.Username)
	if err != nil {
		logrus.Error(err)
		return err, http.StatusInternalServerError
	}
	matching, err := internal.IsPasswordMatching(loginData.Username, loginData.Password,
		playerDetails.Salt, playerDetails.Password)
	if err != nil {
		logrus.Errorf("Error when checking player password: %s", err.Error())
		return err, http.StatusUnauthorized
	}

	if !matching {
		logrus.Warning("Username or password mismatch")
		return errors.New("username or password mismatch"), http.StatusUnauthorized
	}

	logrus.Infof("Successfully logged in user: %s", loginData.Username)
	return nil, http.StatusOK
}
