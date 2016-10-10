
## Simple Redis Mutex Provider for Go



```
import (
	"time"
	"github.com/cloud66/cxthreading/redis"
)
```

```
	// simple example
  
  // arguments: 
  // redisHost --> Redis host address
  // redisNamespace --> Namespace to use within Redis (can be empty string)
  // globalScope --> Sub-namespace for mutex locks from this application
	mutex := redis.NewMutex("localhost:6379", "myRedisNamespace", "myGlobalAppScope")

  // arguments: 
  // localScope --> Sub-sub-namespace for final Redis key identifier
  // timeToWait --> This is the max time to wait before timing out if the lock is not available
  // checkFrequency --> How often to check if this lock is now available
  // functionPointer --> The parameterless function to call if the lock is acquired
	mutex.Synchronise("someLocalScope", 30*time.Second, 2*time.Second, testMethodCallWithoutParam)

  ...
  func testMethodCallWithoutParam() {
	  fmt.Println("called! without param")
  }
```
```
  // with params
	mutex.Synchronise("blah.id1", 60*time.Second, 4*time.Second, func() { testMethodCallWithParam("sample param1") })
  
  ...
  func testMethodCallWithParam(exampleParam string) {
	fmt.Printf("called! exampleParam is: '%s'\n", exampleParam)
  }
```

#### Warning! 
Anonymous method params are not copied on creation! 
(the `Synchronise` function requires parameter-less functions)
```
	paramValue = "sample param2"
	funcPointer := func() { testMethodCallWithParam(paramValue) }
	// Here: will update the value in the pointer!!
	paramValue = "sample params booooo"
	mutex.Synchronise("blah.id2", 30*time.Second, 2*time.Second, funcPointer)
```
To get around this (if it is a problem) you can create your own wrapper struct, initialise it with your required arguments and then call a method on that struct directly.


