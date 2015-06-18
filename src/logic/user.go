package logic

import (
	"fmt"
	. "types"

	"github.com/GameGophers/libs/db"
	log "github.com/GameGophers/nsq-logger"
)

func UserInit(uid int32, sess *Session) error {
	user := &User{
		Id:    uid,
		Name:  fmt.Sprintf("player%v", uid),
		Level: 1,
	}
	sess.User = user

	//TODO init other struct
	return nil
}

func UserLoad(uid int32, sess *Session) error {
	user := &User{}
	if err := db.Redis.Get("users", uid, user); err != nil {
		log.Critical("load user error %v", err)
	}

	//TODO load other struct
	return nil
}
