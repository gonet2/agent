// Strong interactive data should be flush with instant,
// Weak/None interactive data should be use timing.
package main

import (
	"client_handler"
	"time"
	. "types"

	log "github.com/GameGophers/nsq-logger"
	"github.com/fzzy/radix/redis"

	"db"
)

var (
	FLUSH_INTERVAL = 600 //second
	FLUSH_OPS      = 300
)

var (
	_redis_client *redis.Client
)

//需要即时刷入的协议-数据表, 需要map所有表。
var _flush_tbl = map[string]map[string]int8{
	"user_login_req": map[string]int8{"users": 0, "buildings": 0},
}

//--------------------------------------------- flush user data, query tbl data type (strong/weak interactive) to flush
func flush(sess *Session, rcode int16) {
	api := client_handler.RCode[rcode]
	log.Info("flush data user: %s, code: %s, api: %s", sess.UserId, rcode, api)
	if _flush_tbl[api] != nil {
		save2db(sess, _flush_tbl[api])
	}
}

func save2db(sess *Session, tbl map[string]int8) {
	for table, is_instant := range tbl {
		switch table {
		case "users":
			if is_instant == 0 {
				if sess.DirtyCount() < int32(FLUSH_OPS) && (time.Now().Unix()-sess.User.LastSaveTime < int64(FLUSH_INTERVAL)) {
					continue
				}
			}
			if sess.User != nil {
				sess.User.LastSaveTime = time.Now().Unix()
				err := db.Client.Set(table, sess.User.Id, sess.User)
				if err != nil {
					log.Error(err)
				}
				//TODO send dirty key to bgsave, need to do frequency control

			}
			//TODO other table need to save.

		}
	}

	sess.MarkClean()
}
