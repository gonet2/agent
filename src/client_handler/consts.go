package client_handler

const (
	SALT = "DH"
)

const (
	GAME_SRV_HALT    = 0 //不可用
	GAME_SRV_OK      = 1 //正常
	GAME_SRV_CROWDED = 2 //拥挤
	GAME_SRV_FULL    = 3 //满员
)
