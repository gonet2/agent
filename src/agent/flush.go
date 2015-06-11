// Strong interactive data should be flush with instant,
// Weak/None interactive data should be use timing.
package main

import (
	"client_handler"
	"fmt"
	"os"
	"time"
	. "types"

	log "github.com/GameGophers/nsq-logger"
	"github.com/fzzy/radix/redis"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	FLUSH_INTERVAL = 600 //second
	FLUSH_OPS      = 300

	DEFAULT_REDIS_HOST = "127.0.0.1:6379"
	ENV_REDIS_HOST     = "REDIS_HOST"
)

var (
	_redis_client *redis.Client
)

//需要即时刷入的协议-数据表, 需要map所有表。
var _instant_flush_tbl = map[string]map[string]int8{
	"user_login_req": map[string]int8{"users": 0, "buildings": 0},
}

func init() {
	// read redis host
	redis_host := DEFAULT_REDIS_HOST
	if env := os.Getenv(ENV_REDIS_HOST); env != "" {
		redis_host = env
	}
	// start connection to redis
	client, err := redis.Dial("tcp", redis_host)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	_redis_client = client

}

//--------------------------------------------- flush user data, query tbl data type (strong/weak interactive) to flush
func _flush(sess *Session, rcode int16) {
	api := client_handler.RCode[rcode]
	log.Info("flush data user: %s, code: %s, api: %s", sess.UserId, rcode, api)
	if _instant_flush_tbl[api] != nil {
		_save2db(sess, _instant_flush_tbl[api])
	}
}

func _save2db(sess *Session, tbl map[string]int8) {
	is_instant := false
	save_table := make(map[string]int8)
	for k, v := range tbl {
		if v == 1 {
			is_instant = true
			save_table[k] = v
		}

	}

	if is_instant == false {
		if sess.DirtyCount() < int32(FLUSH_OPS) && (time.Now().Unix()-sess.User.LastSaveTime < int64(FLUSH_INTERVAL)) {
			return
		}
		save_table = tbl
	}
	for table, _ := range save_table {
		switch table {
		case "users":
			if sess.User != nil {
				sess.User.LastSaveTime = time.Now().Unix()
				// save to redis.
				bin, err := msgpack.Marshal(sess.User)
				if nil != err {
					log.Critical(err)
				}
				key := fmt.Sprintf("%s:%s", table, sess.User.Id)
				reply, err := _redis_client.Cmd("set", key, bin).Str()
				if err != nil {
					log.Error(err)
				}
				log.Info("flush user:%s reply: %s", sess.User.Id, reply)
				//TODO send dirty key to bgsave

			}
			//TODO other table need to save.

		}
	}

	sess.MarkClean()
}
