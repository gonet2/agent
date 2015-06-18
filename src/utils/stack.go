package utils

import (
	log "github.com/GameGophers/libs/nsq-logger"
	"github.com/davecgh/go-spew/spew"
	"runtime"
)

func PrintPanicStack(extras ...interface{}) {
	if x := recover(); x != nil {
		log.Error(x)
		i := 0
		funcName, file, line, ok := runtime.Caller(i)
		for ok {
			log.Errorf("frame %v:[func:%v,file:%v,line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
			i++
			funcName, file, line, ok = runtime.Caller(i)
		}

		for k := range extras {
			log.Errorf("EXRAS#%v DATA:%v\n", k, spew.Sdump(extras[k]))
		}
	}
}
