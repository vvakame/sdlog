package aelog_test

import (
	"context"
	"fmt"
	"github.com/vvakame/sdlog/aelog"
	"go.opencensus.io/trace"
	"net/http"
	"os"
	"runtime/debug"
)

func Example() {
	_ = os.Setenv("GAE_APPLICATION", "example")
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	// mimic AppEngine environment's request
	r.Header.Add("X-Cloud-Trace-Context", "e5e6be3a33c969fc2c696daf9d969fd9/4901377287092830619;o=1")

	// in your app...
	ctx := context.Background()
	ctx = aelog.WithHTTPRequest(ctx, r)

	aelog.Debugf(ctx, "appengine compat log: %s", "💕 AppEngine")
}

func Example_WithOpenCensus() {
	_ = os.Setenv("GAE_APPLICATION", "example")

	// in your app...
	ctx := context.Background()
	ctx, span := trace.StartSpan(ctx, "foobar")
	defer span.End()

	aelog.Debugf(ctx, "appengine compat log: %s", "💕 AppEngine")
}

func Example_Recover() {
	_ = os.Setenv("GAE_APPLICATION", "example")

	// in your app...
	ctx := context.Background()
	ctx, span := trace.StartSpan(ctx, "foobar")
	defer span.End()
	defer func() {
		err := recover()
		if err != nil {
			// not perfect... but better.
			aelog.Errorf(ctx, "%s\n\n%s", err, string(debug.Stack()))
			panic(err)
		}
	}()

	var obj *http.Request
	// boom!!
	fmt.Println(obj.URL.String())
}
