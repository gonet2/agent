package services

import (
	"testing"
	"time"
)

func TestService(t *testing.T) {
	if _, err := GetService(SERVICE_SNOWFLAKE); err != nil {
		t.Log(err)
	}
	<-time.After(time.Hour)
}
