package slog

import (
	"bytes"
	"errors"
	// "fmt"
	"path"
	"runtime"
	"strconv"
	"time"
)

const (
	FILE_MODE            = 0 //log输出模式
	DEFAULT_ERROR_WORD   = "[E]"
	DEFAULT_WARN_WORD    = "[W]"
	DEFAULT_INFO_WORD    = "[I]"
	DEFAULT_DEBUG_WORD   = "[D]"
	DEFAULT_CALLER_DEPTH = 2
	LEVEL_CUSTOM         = iota
	LEVEL_DEBUG
	LEVEL_INFO
	LEVEL_WARN
	LEVEL_ERROR
)

//logger接口
type ILogger interface {
	InitLogger(configJson string) error
	WriteLogPackage(logPackage *LogPackage) error
	Repaking(log *LogEntity) (bool, error)
	Close() error
}

//单条log的包装实体
type LogEntity struct {
	Msg  string
	When time.Time
}

//log打包输出类
type LogPackage struct {
	Msg            string
	When           time.Time
	LineCount      int
	NeedChangeFile bool
}

//主log处理类
type SLogger struct {
	logger      ILogger
	msgChan     chan *LogEntity
	packageChan chan *LogPackage
	flagChan    chan bool
	writeOver   bool
	level       int
	callerDepth int
}

func NewSLogger(chanSize int64) *SLogger {
	sLogger := &SLogger{
		msgChan:     make(chan *LogEntity, chanSize),
		packageChan: make(chan *LogPackage, chanSize),
		flagChan:    make(chan bool),
		writeOver:   false,
		callerDepth: DEFAULT_CALLER_DEPTH,
	}
	return sLogger
}

//
//action: set log level
//
//params: level int
//return:
//
func (this *SLogger) SetLevel(level int) {
	this.level = level
}

//
//action: set caller depth
//
//params: callerDepath int
//return:
//
func (this *SLogger) SetCallerDepth(callerDepth int) {
	this.callerDepth = callerDepth
}

//
//action:
//
//params:
//return:
//
func (this *SLogger) ConfigLogger(logType int, configJson string) error {
	// fmt.Println("ConfigLogger -> ", logType, " ", configJson)
	switch logType {
	case FILE_MODE:
		this.logger = NewFileLogger()
		err := this.logger.InitLogger(configJson)
		if err != nil {
			return err
		}
	default:
		return errors.New("not support value is \"" + strconv.Itoa(logType) + "\" log mode!")
	}
	go this.writeMsg()
	go this.packageMsg()
	return nil
}

//
//action: output the custom log, please use LEVEL_CUSTOM
//
//params:
//return:
//
func (this *SLogger) WriteMsg(msg string, level int) {
	when := time.Now()
	if level != LEVEL_CUSTOM {
		_, file, line, ok := runtime.Caller(this.callerDepth)
		if !ok {
			file = "???"
			line = 0
		}
		_, filename := path.Split(file)
		msg = when.Format("2006-01-02 15:04:05.000") + " [" + filename + ":" + strconv.FormatInt(int64(line), 10) + "] " + msg
	}
	logEntity := &LogEntity{Msg: msg + "\n", When: when}
	this.msgChan <- logEntity
}

//
//action: Packaging MSG
//
//params:
//return:
//
func (this *SLogger) packageMsg() {
	var buffer bytes.Buffer //Buffer是一个实现了读写方法的可变大小的字节缓冲
	var when time.Time
	var lineCount int
	for {
		select {
		case <-this.flagChan:
			if buffer.Len() <= 0 {
				// fmt.Println("no buffer, get from msgChan")
				entity := <-this.msgChan
				ok, err := this.logger.Repaking(entity)
				this.packageChan <- &LogPackage{Msg: entity.Msg, When: entity.When, LineCount: 1, NeedChangeFile: ok || err != nil}
				buffer.Reset()
				lineCount = 0
				when = entity.When
			} else {
				// fmt.Println("get from buffer")
				this.packageChan <- &LogPackage{Msg: buffer.String(), When: when, LineCount: lineCount}
				buffer.Reset()
				lineCount = 0
			}
			buffer.Reset()
		case entity := <-this.msgChan:
			ok, err := this.logger.Repaking(entity)
			if ok && err == nil {
				// fmt.Println("repaking")
				this.packageChan <- &LogPackage{Msg: buffer.String(), When: when, LineCount: lineCount, NeedChangeFile: true}
				buffer.Reset()
				lineCount = 0
			}
			// fmt.Println("package append msg :", entity)
			buffer.WriteString(entity.Msg)
			lineCount++
			when = entity.When
		}
	}
}

//
//action: write msg
//
//params:
//return:
//
func (this *SLogger) writeMsg() {
	for {
		this.flagChan <- true
		packegeMsg := <-this.packageChan
		if len(packegeMsg.Msg) > 0 {
			// fmt.Println("write msg:", packegeMsg)
			this.logger.WriteLogPackage(packegeMsg)
		}
	}
}

//
//action:
//
//params:
//return:
//
func (this *SLogger) Close() error {
	if this.logger != nil {
		err := this.logger.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

//
//action:
//
//params:
//return:
//
func (this *SLogger) Error(msg string) {
	if this.level > LEVEL_ERROR {
		return
	}
	this.WriteMsg(DEFAULT_ERROR_WORD+" "+msg, LEVEL_ERROR)
}

//
//action:
//
//params:
//return:
//
func (this *SLogger) Warn(msg string) {
	if this.level > LEVEL_WARN {
		return
	}
	this.WriteMsg(DEFAULT_WARN_WORD+" "+msg, LEVEL_WARN)
}

//
//action:
//
//params:
//return:
//
func (this *SLogger) Info(msg string) {
	if this.level > LEVEL_INFO {
		return
	}
	this.WriteMsg(DEFAULT_INFO_WORD+" "+msg, LEVEL_INFO)
}

//
//action:
//
//params:
//return:
//
func (this *SLogger) Debug(msg string) {
	if this.level > LEVEL_DEBUG {
		return
	}
	this.WriteMsg(DEFAULT_DEBUG_WORD+" "+msg, LEVEL_DEBUG)
}
