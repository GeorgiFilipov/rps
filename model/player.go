package model

// PlayerLoginRequest data that is used to attempt a player login
type PlayerLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Player stores and manages all player data. The player registers with a port.
type Player struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
	Balance  int    `json:"balance"`
}

// PlayerRegistrationRequest data required to register a player
type PlayerRegistrationRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Deposit  int    `json:"deposit" binding:"required"`
}
