package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
)

const k = 1 << 10

func PanicRecover(r interface{}) error {
	loggerStd := log.New(os.Stdout, "", log.LstdFlags)

	if r != nil {
		buf := make([]byte, 4*k)
		n := runtime.Stack(buf, false)
		loggerStd.Printf("[Recovery] panic recovered:\n%s\n%s\n", r, buf[:n])
		logrus.WithFields(logrus.Fields{
			"stack": string(buf[:n]),
		}).Errorf("[Recovery] panic recovered:%#v", r)

		var err error
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = fmt.Errorf("%v", r)
		}

		return err
	}

	return nil
}
