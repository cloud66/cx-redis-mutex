package main

import (
	"fmt"
	"time"

	"github.com/cloud66/cxlogger"
	"github.com/cloud66/cxthreading/redis"
)

func main() {
	// some simple tests
	cxlogger.Initialize("STDOUT", "debug")
	cxlogger.Log.Debug("starting")
	mutex := redis.NewMutex("localhost:6379", "redisGoMutex", "ironmount")
	cxlogger.Log.Debug("instance created")

	mutex.Synchronise("action_save", 30*time.Second, 2*time.Second, testMethodCallWithoutParam)

	paramValue := "sample param1"
	mutex.Synchronise("vault.id2", 30*time.Second, 2*time.Second, func() { testMethodCallWithParam(paramValue) })

	paramValue = "sample param2"
	funcPointer := func() { testMethodCallWithParam(paramValue) }
	// WARNING: will update the value in the pointer!!
	paramValue = "sample params booooo"
	mutex.Synchronise("vault.id2", 30*time.Second, 2*time.Second, funcPointer)

}

func testMethodCallWithoutParam() {
	fmt.Println("called! without param")
	time.Sleep(5 * time.Second)

}

func testMethodCallWithParam(exampleParam string) {
	fmt.Printf("called! exampleParam is: '%s'\n", exampleParam)
	time.Sleep(5 * time.Second)

}
