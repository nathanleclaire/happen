package happen

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/Sirupsen/logrus"
)

type Happening struct {
	begin, end time.Time
}

type Actioner struct {
	Err  error
	Func string
}

func (a *Actioner) Log() {
	End(a.Func)

	if a.Err != nil {
		logrus.Error(a.Err)
		return
	}

	timespent, err := Duration(a.Func)
	if err != nil {
		logrus.Error(err)
		return
	}

	// TODO:
	// (1) Something more general, but preferably without a lot of pomp and
	// circumstance when used.
	//
	// (2) Make goroutine-safe. Timed actions in seprate goroutines with
	// TimeMe() should not interfere with each other.
	logrus.WithFields(logrus.Fields{
		"name":     a.Func,
		"duration": timespent,
	}).Info("Finished timing " + a.Func)
}

var happenings = map[string]*Happening{}

func Begin(key string) error {
	if _, ok := happenings[key]; ok {
		return fmt.Errorf("%q already started happening", key)
	}
	happenings[key] = &Happening{begin: time.Now()}
	return nil
}

func End(key string) error {
	happening, ok := happenings[key]
	if !ok {
		return fmt.Errorf("%q did not start happening yet", key)
	}
	happening.end = time.Now()
	happenings[key] = happening
	return nil
}

// Duration returns how much time occurred in the happening. It also clears
// the key.
func Duration(key string) (time.Duration, error) {
	happening, ok := happenings[key]
	if !ok {
		return 0, fmt.Errorf("%q didn't happen", key)
	}

	dur := happening.end.Sub(happening.begin)
	delete(happenings, key)
	return dur, nil
}

// TimeMe times a particular function call as a happening and returns a struct
// with actions that can be deferred to do something (e.g., log the timer) once
// the timing is complete.
func TimeMe() *Actioner {
	actioner := &Actioner{}
	pc, _, _, ok := runtime.Caller(1)
	caller := runtime.FuncForPC(pc)
	if !ok || caller == nil {
		actioner.Err = errors.New("Unable to recover caller information")
		return actioner
	}
	actioner.Func = caller.Name()
	Begin(actioner.Func)
	return actioner
}
