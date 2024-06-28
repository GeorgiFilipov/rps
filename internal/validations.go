package internal

import (
	"errors"
	"github.com/sirupsen/logrus"
	"main/config"
)

func ValidateMinimumPlayerDeposit(deposit int) error {
	// ... user balance is above minimum
	if valid := deposit > config.Settings.MinimumDeposit; !valid {
		logrus.Errorf("Invalid balance for player: %d , must be %d or more", deposit, config.Settings.MinimumDeposit)
		return errors.New("balance is invalid")
	}

	return nil
}

func ValidatePlayerUsername(username string) error {
	// ... username is not too short
	if valid := len(username) > config.Settings.MinimumNameLength; !valid {
		logrus.Errorf("Username is too short, %s, needs to be at least %d characaters long", username, config.Settings.MinimumNameLength)
		return errors.New("username is too short")
	}

	// ... username is not too long
	if valid := len(username) < config.Settings.MaximumNameLength; !valid {
		logrus.Errorf("Username is too long, %s, needs to be at less than %d characaters long", username, config.Settings.MinimumNameLength)
		return errors.New("username is too long")
	}

	return nil
}

func ValidatePlayerPassword(password string) error {
	if valid := len(password) > config.Settings.MinimumPasswordLength; !valid {
		logrus.Errorf("Password is too short: %s", password)
		return errors.New("password is too short")
	}

	return nil
}

func IsPasswordMatching(username, password, salt, hashedPassword string) (bool, error) {
	// Hash password with salt for the found username
	inputPassHash, err := HashPassword(password, salt)
	if err != nil {
		logrus.Errorf("Error hashing password: %v", err)
		return false, err
	}

	// Compare both passwords
	if inputPassHash == hashedPassword {
		logrus.Infof("Password matches for username %s", username)
		return true, nil
	}

	logrus.Errorf("Password mismatched for user: %s", username)
	return false, errors.New("invalid password or username")
}
