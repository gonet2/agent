// Strong interactive data should be flush with instant,
// Weak/None interactive data should be use timing.
package main

import (
	"client_handler"
	"fmt"
	"os"
	"time"
	. "types"

	"gopkg.in/vmihailenco/msgpack.v2"

	log "github.com/GameGophers/nsq-logger"
	"github.com/fzzy/radix/redis"
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
var _flush_tbl = map[string]map[string]int8{
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
				//TODO send dirty key to bgsave, need to do frequency control

			}
			//TODO other table need to save.

		}
	}

	sess.MarkClean()
}
