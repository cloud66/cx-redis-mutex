package redis

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloud66/cxlogger"
	"github.com/garyburd/redigo/redis"
)

// Mutex containing mutex methods
type Mutex struct {
	RedisHost      string
	RedisNamespace string
	GlobalScope    string
	pool           *redis.Pool
	acquireScript  *redis.Script
}

func newPool(redisHost string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

// NewMutex creates a new redisMutex instance
func NewMutex(redisHost string, redisNamespace string, globalScope string) Mutex {
	acquireRaw := "return redis.call('setnx', KEYS[1], 1) == 1 and redis.call('expire', KEYS[1], KEYS[2]) and 1 or 0"
	return Mutex{
		RedisHost:      redisHost,
		RedisNamespace: redisNamespace,
		GlobalScope:    globalScope,
		pool:           newPool(redisHost),
		acquireScript:  redis.NewScript(2, acquireRaw),
	}
}

// Synchronise will synchronise calls to the requested function
// ie. redisMutex.Synchronise("test", 1*time.Minute, 10*time.Second, someMethodCall)
func (r Mutex) Synchronise(localScope string, waitExpiration time.Duration, checkFrequency time.Duration, function func()) error {
	// get the lock
	cxlogger.Log.Debug("about to acquire lock")
	err := r.acquire(localScope, waitExpiration, checkFrequency)
	if err != nil {
		return err
	}
	// defer release of the lock
	defer r.release(localScope)
	function()
	return nil
}

func (r Mutex) acquire(localScope string, waitExpiration time.Duration, checkFrequency time.Duration) error {
	var expired bool
	mutexKey := fmt.Sprintf("%s.%s.%s", r.RedisNamespace, r.GlobalScope, localScope)
	currentTime := time.Now()
	finalTime := currentTime.Add(waitExpiration)
	expired = finalTime.Before(currentTime)
	if expired == true {
		return errors.New("Wait expired without lock")
	}

	cxlogger.Log.Debug("getting redis connection")
	redisConn := r.pool.Get()
	defer redisConn.Close()

	for {
		// attempt to get the mutex
		cxlogger.Log.Debug("running script")
		reply, err := r.acquireScript.Do(redisConn, mutexKey, waitExpiration.Seconds())
		if err != nil {
			return fmt.Errorf("Redis action failed. %s", err.Error())
		}
		cxlogger.Log.Debug("Got reply!")
		cxlogger.Log.Debug(fmt.Sprintf("%v", reply))
		if reply == int64(1) {
			// we have a the mutex
			cxlogger.Log.Debug("got it!")
			break
		}
		cxlogger.Log.Debug("didn't get it, waiting!")

		// wait for lock to become available and try again
		time.Sleep(checkFrequency)
		currentTime = time.Now()
		expired = finalTime.Before(currentTime)
		if expired == true {
			cxlogger.Log.Debug("didn't get it and now it has expired")
			break
		}
	}

	if expired == true {
		return errors.New("Wait expired without lock")
	}
	return nil
}

func (r Mutex) release(localScope string) error {
	cxlogger.Log.Debug("releasing scope")
	mutexKey := fmt.Sprintf("%s.%s.%s", r.RedisNamespace, r.GlobalScope, localScope)
	redisConn := r.pool.Get()
	defer redisConn.Close()
	_, err := redisConn.Do("DEL", mutexKey)
	return err
}
