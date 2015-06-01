package services

import (
	"testing"
	"time"
)

func TestService(t *testing.T) {
	if _, err := _default_pool.get_snowflake(); err != nil {
		t.Log(err)
	}
	<-time.After(time.Hour)
}
