package slog

import (
	"bufio"
	"encoding/json"
	"errors"
	// "fmt"
	"os"
	"time"
)

const (
	DefaultMaxLine = 100000
	DeFaultMaxSize = 512 * 1024 * 1024
	DefaultMaxDay  = 7
)

type FileLogger struct {
	FilePath       string `json:filePath`
	FileName       string `json:"fileName"`
	MaxLine        int64  `json:"maxLine"`
	currentLine    int64
	IsAllowMaxLine bool  `json:"isAllowMaxLine"`
	MaxSize        int64 `json:"maxSize"`
	currentSize    int64
	IsAllowMaxSize bool `json:"isAllowMaxSize"`
	MaxDay         int  `json:"maxDay"`
	preTime        time.Time
	IsAllowMaxDay  bool `json:"isAllowMaxDay"`
	logFile        *os.File
	writer         *bufio.Writer
}

func NewFileLogger() *FileLogger {
	fileLogger := &FileLogger{
		IsAllowMaxSize: true,
		MaxSize:        DeFaultMaxSize,
		currentLine:    0,
		IsAllowMaxLine: true,
		MaxLine:        DefaultMaxLine,
		currentSize:    0,
		IsAllowMaxDay:  true,
		MaxDay:         DefaultMaxDay,
		preTime:        time.Now(),
	}
	return fileLogger
}

//
//action: init logger
//
//params:
//return:
//
func (this *FileLogger) InitLogger(configJson string) error {
	var err error
	if len(configJson) > 0 {
		err = json.Unmarshal([]byte(configJson), this)
		if err != nil {
			// fmt.Println(err.Error())
			return err
		}
	} else {
		return errors.New("configJson cannot be nil")
	}
	if len(this.FileName) <= 0 {
		return errors.New("configJson must hava filename")
	}

	fi, err := os.Stat(this.FilePath)

	if err != nil || !fi.IsDir() {
		err = os.MkdirAll(this.FilePath, os.ModeDir|0777)
		if err != nil {
			// fmt.Println(err.Error())
			return err
		}
	}
	return this.createLogFile()
}

//
//action:
//
//params:
//return:
//
func (this *FileLogger) createLogFile() error {
	var err error
	this.logFile, err = os.OpenFile(this.FilePath+"/"+this.FileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		// fmt.Println(err.Error())
		return err
	}
	this.writer = bufio.NewWriter(this.logFile)
	return nil
}

//
//action:
//
//params:
//return:
//
func (this *FileLogger) nextLogFile() error {
	this.Close()
	fileName := this.FilePath + "/" + this.FileName
	newFileName := fileName + "." + time.Now().Format("2006-01-02_15-04-05.000000000")
	err := os.Rename(fileName, newFileName)
	if err != nil {
		// fmt.Println(err.Error())
		return err
	}
	this.createLogFile()
	return nil
}

//
//action:
//
//params:
//return:
//
func (this *FileLogger) WriteLogPackage(logPackage *LogPackage) error {
	this.currentLine += int64(logPackage.LineCount)
	_, err := this.writer.WriteString(logPackage.Msg)
	if err != nil {
		// fmt.Println(err.Error())
		return err
	}
	if logPackage.NeedChangeFile {
		this.nextLogFile()
	}
	return nil
}

//
//action:
//
//params:
//return:
//
func (this *FileLogger) Repaking(log *LogEntity) (bool, error) {
	result := false
	if this.IsAllowMaxDay && log.When.Sub(this.preTime) > 7*24*time.Hour {
		// fmt.Println("day------------------------")
		result = true
	}
	if !result && this.IsAllowMaxLine {
		if this.currentLine+1 > this.MaxLine {
			// fmt.Println("line------------------------")
			result = true
		}
		this.currentLine++
	}
	if !result && this.IsAllowMaxSize {
		size := len([]byte(log.Msg))
		if this.currentSize+int64(size) > this.MaxSize {
			// fmt.Println("size------------------------")
			result = true
		}
		this.currentSize += int64(size)
	}

	if result {
		this.currentSize = 0
		this.currentLine = 0
		this.preTime = log.When
	}
	return result, nil
}

//
//action:
//
//params:
//return:
//
func (this *FileLogger) Close() error {
	if this.logFile != nil {
		this.writer.Flush()
		err := this.logFile.Close()
		if err != nil {
			// fmt.Println(err.Error())
			return err
		}
	}
	return nil
}
