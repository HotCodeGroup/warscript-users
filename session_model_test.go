package main

import (
	"testing"
	"time"

	"github.com/HotCodeGroup/warscript-utils/utils"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// newTestRedis returns a redis.Cmdable.
func newTestRedis() *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	return redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
}

func TestSetOK(t *testing.T) {
	rediCli = newTestRedis()
	Sessions = &SessionConn{}

	s := &Session{
		Token:   "kek",
		Payload: []byte{1, 2, 3},
	}
	if err := Sessions.Set(s); err != nil {
		t.Errorf("TestSetOK got unexpected error: %v", err)
	}
}

func TestSetErr(t *testing.T) {
	rediCli = redis.NewClient(&redis.Options{})
	Sessions = &SessionConn{}

	s := &Session{
		Token:   "kek",
		Payload: []byte{1, 2, 3},
	}

	err := Sessions.Set(s)
	if errors.Cause(err) != utils.ErrInternal {
		t.Errorf("TestSetErr got unexpected error: %v, expected: %v", err, utils.ErrInternal)
	}
}

func TestDeleteOK(t *testing.T) {
	rediCli = newTestRedis()
	Sessions = &SessionConn{}

	s := &Session{
		Token:   "kek",
		Payload: []byte{1, 2, 3},
	}
	if err := Sessions.Delete(s); err != nil {
		t.Errorf("TestDeleteOK got unexpected error: %v", err)
	}
}

func TestDeleteErr(t *testing.T) {
	rediCli = redis.NewClient(&redis.Options{})
	Sessions = &SessionConn{}

	s := &Session{
		Token:   "kek",
		Payload: []byte{1, 2, 3},
	}

	err := Sessions.Delete(s)
	if errors.Cause(err) != utils.ErrInternal {
		t.Errorf("TestDeleteErr got unexpected error: %v, expected: %v", err, utils.ErrInternal)
	}
}

func TestGetSessionModelOK(t *testing.T) {
	rediCli = newTestRedis()
	Sessions = &SessionConn{}

	rediCli.Set("kek", "lol", time.Minute)

	if _, err := Sessions.GetSession("kek"); err != nil {
		t.Errorf("TestGetSessionModelOK got unexpected error: %v", err)
	}
}

func TestGetSessionModelErr(t *testing.T) {
	rediCli = redis.NewClient(&redis.Options{})
	Sessions = &SessionConn{}

	_, err := Sessions.GetSession("kek")
	if errors.Cause(err) != utils.ErrInternal {
		t.Errorf(" TestGetSessionModel got unexpected error: %v, expected: %v", err, utils.ErrInternal)
	}
}
