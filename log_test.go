package slog

import (
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	t.Log("test start")
	runtime.GOMAXPROCS(runtime.NumCPU())
	sLogger := NewSLogger(10000)
	err := sLogger.ConfigLogger(FILE_MODE, `{"filePath":"E:/GOPATH/src/github.com/sipt/slog/logs","fileName":"action.log","maxLine":10,"isAllowMaxLine":false,"isAllowMaxSize":true,"maxSize":1000000,"isAllowMaxDay":false,"maxDay":7}`)
	if err != nil {
		panic(err)
	}

	end := make(chan bool, 1000)
	for i := 0; i < 1000; i++ {
		outputLog(sLogger, i, end)
	}
	time.Sleep(10 * time.Second)
	sLogger.Close()
	t.Log("test end")

}

func outputLog(sLogger *SLogger, number int, end chan bool) {
	for i := 0; i < 1000; i++ {
		sLogger.Info("thread " + strconv.Itoa(number) + ", this is a log!!!!")
	}
	end <- true
}

func outputShotLog(sLogger *SLogger, number int, end chan bool) {
	for i := 0; i < 1000; i++ {
		sLogger.Info(strconv.Itoa(number))
	}
	end <- true
}
