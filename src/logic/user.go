package logic

import (
	"fmt"
	. "types"

	"gopkg.in/mgo.v2/bson"

	"github.com/GameGophers/libs/db"
)

func AuthInit(uid int32, sess *Session) error {
	auth := &Auth{
		Id:   uid,
		Name: fmt.Sprintf("player%v", uid),
	}
	sess.Auth = auth

	//TODO init other struct
	return nil
}

func UserLoad(uid int32, sess *Session) error {
	auth := &Auth{}
	ms, c := db.C("auth")

	//TODO load other struct
	return nil
}

//---------------------------------------------------------- update/insert a user
func Set(user *User) bool {
	ms, c := C(COLLECTION)
	defer ms.Close()

	info, err := c.Upsert(bson.M{"id": user.Id}, user)
	if err != nil {
		utils.ERR(COLLECTION, "Set", info, err, user)
		return false
	}

	return true
}

//---------------------------------------------------------- find by id
func FindById(id int32) *User {
	ms, c := C(COLLECTION)
	defer ms.Close()

	user := &User{}
	err := c.Find(bson.M{"id": id}).One(user)
	if err != nil {
		utils.WARN(COLLECTION, "FindById", err, id)
		return nil
	}

	return user
}
