package util

import (
	"fmt"
	"project/MyRedis/util"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	defaultExpiration, _ := time.ParseDuration("0.5h")
	gcInterval, _ := time.ParseDuration("3s")
	c := util.NewCache(defaultExpiration, gcInterval)
	k1 := 12121
	expiration := "30s"
	c.Set("k1", k1, expiration)
	v, found := c.Get("k1")
	if !found {
		fmt.Println(found)
	} else {
		fmt.Println(v)
	}
}
