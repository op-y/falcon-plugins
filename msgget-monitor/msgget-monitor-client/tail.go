/*
* tail.go - file agent data structure and funtions to tail file
*
* history
* --------------------
* 2017/8/18, by Ye Zhiqin, create
* 2017/9/30, by Ye Zhiqin, modify
* 2018/1/3,  by Ye Zhiqin, modify
* 2018/1/11, by Ye Zhiqin, modify
*
* DESCRIPTION
* This file contains the definition of file agent
* and the functions to tail log file
 */

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/fsnotify/fsnotify"
)

type FileAgent struct {
	Filename  string
	Delimiter string
	Inotify   bool
	SelfCheck bool

	File         *os.File
	FileInfo     os.FileInfo
	LastOffset   int64
	UnchangeTime int

	Data *AgentData
}

type AgentData struct {
	Endpoint    string
	Tags        string
	CounterType string
	Step        int64
	Pattern     string

	TsStart  int64
	TsEnd    int64
	TsUpdate int64

	ErrorCnt  int64
	ReqCnt    int64
	Req200Cnt int64
	Req499Cnt int64
	Req500Cnt int64
	Req502Cnt int64
	Req504Cnt int64

	IpSet *StringSet
	IdSet *StringSet
}

/*
* Report - push and update data after a period passed
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   - ts: timestamp
*
* RETURNS:
*   No paramter
 */
func (fa *FileAgent) Report(ts time.Time) {
	if fa.SelfCheck {
		Push2Falcon(config.FalconAgent, fa.Data)
	}

	Push2Redis(config.RedisServer, fa.Data)

	log.Printf("=====Report[%d]=====\nError:%d Total:%d 200:%d 499:%d 500:%d 502:%d 504:%d ip:%d id:%d",
		fa.Data.TsStart,
		fa.Data.ErrorCnt,
		fa.Data.ReqCnt,
		fa.Data.Req200Cnt,
		fa.Data.Req499Cnt,
		fa.Data.Req500Cnt,
		fa.Data.Req502Cnt,
		fa.Data.Req504Cnt,
		fa.Data.IpSet.Len(),
		fa.Data.IdSet.Len())

	// update value
	fa.Data.ErrorCnt = 0
	fa.Data.ReqCnt = 0
	fa.Data.Req200Cnt = 0
	fa.Data.Req499Cnt = 0
	fa.Data.Req500Cnt = 0
	fa.Data.Req502Cnt = 0
	fa.Data.Req504Cnt = 0
	fa.Data.IpSet.Clear()
	fa.Data.IdSet.Clear()

	//update timestamp
	fa.Data.TsStart += fa.Data.Step
	fa.Data.TsEnd += fa.Data.Step
	fa.Data.TsUpdate = ts.Unix()
}

/*
* Timeup - the process after a period passed
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   No paramter
 */
func (fa *FileAgent) Timeup() {
	ts := time.Now()
	if ts.Unix()-fa.Data.TsStart >= fa.Data.Step {
		fa.Report(ts)
	}
}

/*
* MatchLine - process each line of log
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   - line: a line of log file
*
* RETURNS:
*   No paramter
 */
func (fa *FileAgent) MatchLine(line []byte) {
	// convert one request log to a json struct
	contents, err := ioutil.ReadAll(bytes.NewReader(line))
	if err != nil {
		log.Printf("convert line <%s> to *Reader FAIL: %s", line, err.Error())
		fa.Data.ErrorCnt += 1
		return
	}

	json, err := simplejson.NewJson(contents)
	if err != nil {
		log.Printf("convert json string FAIL: %s", err.Error())
		fa.Data.ErrorCnt += 1
		return
	}

	// get request field, return if uri is not /msg/get.htm
	request, err := json.Get("request").String()
	if err != nil {
		fa.Data.ErrorCnt += 1
		log.Printf("get request field in line FAIL: %s", err.Error())
		return
	}

	if ok := MatchRequest(request, fa.Data.Pattern); !ok {
		return
	}

	// process status, ip, shopID info
	status, err := json.Get("status").String()
	if err != nil {
		fa.Data.ErrorCnt += 1
		log.Printf("get status field in line FAIL: %s", err.Error())
		return
	}
	switch status {
	case "200":
		fa.Data.Req200Cnt++
		fa.Data.ReqCnt++
	case "499":
		fa.Data.Req499Cnt++
		fa.Data.ReqCnt++
	case "500":
		fa.Data.Req500Cnt++
		fa.Data.ReqCnt++
	case "502":
		fa.Data.Req502Cnt++
		fa.Data.ReqCnt++
	case "504":
		fa.Data.Req504Cnt++
		fa.Data.ReqCnt++
	default:
		fa.Data.ReqCnt++
	}

	ip, err := json.Get("remote_addr").String()
	if err != nil {
		fa.Data.ErrorCnt += 1
		log.Printf("get remote_addr field in line FAIL: %s", err.Error())
		return
	}
	fa.Data.IpSet.Add(ip)

	body, err := json.Get("request_body").String()
	if err != nil {
		fa.Data.ErrorCnt += 1
		log.Printf("get request_body field in line FAIL: %s", err.Error())
		return
	}
	ok, id := GetID(body)
	if !ok {
		fa.Data.ErrorCnt += 1
		log.Printf("get shopID from request body FAIL")
		return
	}
	fa.Data.IdSet.Add(id)

	return
}

/*
* ReadRemainder - reading new bytes of log file
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   nil: succeed
*   error: fail
 */
func (fa *FileAgent) ReadRemainder() error {
	tailable := fa.FileInfo.Mode().IsRegular()
	size := fa.FileInfo.Size()

	// LastOffset less then new size. Maybe the file has been truncated.
	if tailable && fa.LastOffset > size {
		// seek the cursor to the header of new file
		offset, err := fa.File.Seek(0, os.SEEK_SET)
		if err != nil {
			log.Printf("file %s seek FAIL: %s", fa.Filename, err.Error())
			return err
		}
		if offset != 0 {
			log.Printf("offset is not equal 0")
		}
		fa.LastOffset = 0

		return nil
	}

	bufsize := size - fa.LastOffset
	if bufsize == 0 {
		return nil
	}
	data := make([]byte, bufsize)
	readsize, err := fa.File.Read(data)

	if err != nil && err != io.EOF {
		log.Printf("file %s read FAIL: %s", err.Error())
		return err
	}
	if readsize == 0 {
		log.Printf("file %s read 0 data", fa.Filename)
		return nil
	}

	if fa.Delimiter == "" {
		fa.Delimiter = "\n"
	}
	sep := []byte(fa.Delimiter)
	lines := bytes.SplitAfter(data, sep)
	length := len(lines)

	for idx, line := range lines {
		// just process entire line with the delimiter
		if idx == length-1 {
			backsize := len(line)
			movesize := readsize - backsize

			_, err := fa.File.Seek(-int64(backsize), os.SEEK_CUR)
			if err != nil {
				log.Printf("seek file %s FAIL: %s", fa.Filename, err.Error())
				return err
			}
			fa.LastOffset += int64(movesize)

			break
		}
		fa.MatchLine(line)
	}
	return nil
}

/*
* TailWithCheck - tail log file in a loop
*
* PARAMS:
*   - fa: file agent
*   - finish: a channel to receiver stop signal
*
* RETURNS:
*   No return value
 */
func TailWithCheck(fa *FileAgent, finish <-chan bool) {
	log.Printf("agent for %s is launching...", fa.Filename)

	// create one second ticker
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

TAIL:
	for {
		select {
		case <-finish:
			if fa.File != nil {
				if err := fa.File.Close(); err != nil {
					log.Printf("file closing FAIL: %s", err.Error())
				}
			}
			break TAIL
		case <-ticker.C:
			fa.Timeup()
		default:
			fa.TryReading()
			time.Sleep(time.Millisecond * 250)
		}
	}

	wg.Done()
	log.Printf("wg: %v", wg)
	log.Printf("agent for %s is exiting...", fa.Filename)
}

/*
* TryReading - reading log file
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   No return value
 */
func (fa *FileAgent) TryReading() {
	if fa.File == nil {
		log.Printf("file %s is nil", fa.Filename)
		if err := fa.FileRecheck(); err != nil {
			log.Printf("file recheck FAIL: %s", err.Error())
		}
		return
	}

	if !fa.IsChanged() {
		if fa.UnchangeTime >= MAX_UNCHANGED_TIME {
			fa.FileRecheck()
		}
		return
	}

	fa.ReadRemainder()
}

/*
* TailWithInotify - trace log file
*
* PARAMS:
*   - fa: file agent
*   - finish: a channel to receiver stop signal
*
* RETURNS:
*   No return value
 */
func TailWithInotify(fa *FileAgent, chanFinish <-chan bool) {
	fmt.Printf("agent for %s is starting...\n", fa.Filename)

	// create one second ticker
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// craete file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("watcher creating FAIL: %s", err.Error())
		wg.Done()
		log.Printf("agent for %s is exiting...", fa.Filename)
		return
	}
	defer watcher.Close()

	// add parent directory to watcher
	dir := path.Dir(fa.Filename)
	if err := watcher.Add(dir); err != nil {
		log.Printf("add %s to watcher FAIL: %s", dir, err.Error())
		wg.Done()
		log.Printf("agent for %s is exiting...", fa.Filename)
		return
	}

	// open file and initialize file agent
	if err := fa.FileOpen(); err != nil {
		log.Printf("file $s open FAIL when agent initializing: %s", fa.Filename, err.Error())
	}

	// trace file in this loop
TRACE:
	for {
		select {
		case <-chanFinish:
			if err := watcher.Remove(dir); err != nil {
				log.Printf("watcher file removing FAIL: %s", err.Error())
			}
			if fa.File != nil {
				if err := fa.File.Close(); err != nil {
					log.Printf("file closing FAIL: %s", err.Error())
				}
			}
			break TRACE
		case event := <-watcher.Events:
			if event.Name == fa.Filename {
				// WRITE event
				if 2 == event.Op {
					if err := fa.FileInfoUpdate(); err != nil {
						log.Printf("file %s stat FAIL", fa.Filename)
						if fa.UnchangeTime > MAX_UNCHANGED_TIME {
							fa.FileReopen()
						}
						continue
					}

					if err := fa.ReadRemainder(); err != nil {
						log.Printf("file %s reading FAIL, recheck it", fa.Filename)
						fa.FileReopen()
					}
				}
				// CREATE event
				if 1 == event.Op {
					fmt.Printf("fa %s, watch %s receive event CREATE\n", fa.Filename, event.Name)
					fa.FileReopen()
				}
				// REMOVE/RENAME event
				if 4 == event.Op || 8 == event.Op {
					fmt.Printf("fa %s, watch %s receive event REMOVE|RENAME\n", fa.Filename, event.Name)
					fa.FileClose()
				}
				// CHMOD event
				if 16 == event.Op {
					fmt.Printf("fa %s, watch %s receive event CHMOD\n", fa.Filename, event.Name)
				}
			}
		case err := <-watcher.Errors:
			log.Printf("%s receive error %s", fa.Filename, err.Error())
		case <-ticker.C:
			fa.Timeup()
		default:
			time.Sleep(time.Millisecond * 250)
		}
	}

	wg.Done()

	fmt.Printf("agent for %s is exiting...\n", fa.Filename)
}

/*
* FileOpen - open file while file agent initializing
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   nil, if succeed
*   error, if fail
 */
func (fa *FileAgent) FileOpen() error {
	// close old file
	if fa.File != nil {
		if err := fa.File.Close(); err != nil {
			log.Printf("file closing FAIL: %s", err.Error())
		}
	}

	fa.File = nil
	fa.FileInfo = nil

	// open new file
	filename := fa.Filename

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("file %s open FAIL: %s", fa.Filename, err.Error())
		return err
	}

	fileinfo, err := file.Stat()
	if err != nil {
		log.Printf("file %s stat FAIL: %s", fa.Filename, err.Error())
		return err
	}

	fmt.Printf("file %s is open\n", fa.Filename)

	fa.File = file
	fa.FileInfo = fileinfo
	fa.LastOffset = 0
	fa.UnchangeTime = 0

	// seek the cursor to the end of new file
	_, err = fa.File.Seek(fa.FileInfo.Size(), os.SEEK_SET)
	if err != nil {
		log.Printf("seek file %s FAIL: %s", fa.Filename, err.Error())
	}
	fa.LastOffset += fa.FileInfo.Size()

	// initialize agent data
	now := time.Now()
	minute := now.Format("200601021504")

	tsNow := now.Unix()
	tsStart := tsNow

	start, err := time.ParseInLocation("20060102150405", minute+"00", now.Location())
	if err != nil {
		log.Printf("timestamp setting FAIL: %s", err.Error())
	} else {
		tsStart = start.Unix()
	}

	fa.Data.TsStart = tsStart
	fa.Data.TsEnd = tsStart + fa.Data.Step - 1
	fa.Data.TsUpdate = tsNow
	fa.Data.ErrorCnt = 0
	fa.Data.ReqCnt = 0
	fa.Data.Req200Cnt = 0
	fa.Data.Req499Cnt = 0
	fa.Data.Req500Cnt = 0
	fa.Data.Req502Cnt = 0
	fa.Data.Req504Cnt = 0
	fa.Data.IpSet = NewStringSet()
	fa.Data.IdSet = NewStringSet()

	return nil
}

/*
* FileInfoUpdate - stat file when WRITE
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   nil, if succeed
*   error, if fail
 */
func (fa *FileAgent) FileInfoUpdate() error {
	fileinfo, err := fa.File.Stat()
	if err != nil {
		log.Printf("file %s stat FAIL: %s", fa.Filename, err.Error())
		fa.UnchangeTime += 1
		return err
	}

	fa.FileInfo = fileinfo
	fa.UnchangeTime = 0
	return nil
}

/*
* FileClose - close file when REMOVE/RENAME
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   nil, if succeed
*   error, if fail
 */
func (fa *FileAgent) FileClose() error {
	// close old file
	if fa.File != nil {
		if err := fa.File.Close(); err != nil {
			log.Printf("file closing FAIL: %s", err.Error())
			return err
		}
	}

	// clear file info
	fa.File = nil
	fa.FileInfo = nil
	fa.LastOffset = 0
	fa.UnchangeTime = 0

	// clear data
	fa.Data.TsStart = 0
	fa.Data.TsEnd = 0
	fa.Data.TsUpdate = 0
	fa.Data.ErrorCnt = 0
	fa.Data.ReqCnt = 0
	fa.Data.Req200Cnt = 0
	fa.Data.Req499Cnt = 0
	fa.Data.Req500Cnt = 0
	fa.Data.Req502Cnt = 0
	fa.Data.Req504Cnt = 0
	fa.Data.IpSet = nil
	fa.Data.IdSet = nil

	return nil
}

/*
* FileReopen - reopen file when CREATE/ERROR
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   nil: succeed
*   error: fail
 */
func (fa *FileAgent) FileReopen() error {
	// close old file
	if fa.File != nil {
		if err := fa.File.Close(); err != nil {
			log.Printf("file closing FAIL: %s", err.Error())
		}
	}

	fa.File = nil
	fa.FileInfo = nil

	// open new file
	filename := fa.Filename

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("file %s opening FAIL: %s", fa.Filename, err.Error())
		return err
	}

	fileinfo, err := file.Stat()
	if err != nil {
		log.Printf("file %s stat FAIL: %s", fa.Filename, err.Error())
		return err
	}

	log.Printf("file %s recheck ok, it is a new file", fa.Filename)

	fa.File = file
	fa.FileInfo = fileinfo
	fa.LastOffset = 0
	fa.UnchangeTime = 0

	// seek the cursor to the start of new file
	_, err = fa.File.Seek(0, os.SEEK_SET)
	if err != nil {
		log.Printf("seek file %s FAIL: %s", fa.Filename, err.Error())
	}

	// initialize agent data
	now := time.Now()
	minute := now.Format("200601021504")

	tsNow := now.Unix()
	tsStart := tsNow

	start, err := time.ParseInLocation("20060102150405", minute+"00", now.Location())
	if err != nil {
		log.Printf("timestamp setting FAIL: %s", err.Error())
	} else {
		tsStart = start.Unix()
	}

	fa.Data.TsStart = tsStart
	fa.Data.TsEnd = tsStart + fa.Data.Step - 1
	fa.Data.TsUpdate = tsNow
	fa.Data.ErrorCnt = 0
	fa.Data.ReqCnt = 0
	fa.Data.Req200Cnt = 0
	fa.Data.Req499Cnt = 0
	fa.Data.Req500Cnt = 0
	fa.Data.Req502Cnt = 0
	fa.Data.Req504Cnt = 0
	fa.Data.IpSet = NewStringSet()
	fa.Data.IdSet = NewStringSet()

	return nil
}

/*
* FileRecheck - recheck the file for file agent
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   nil, if succeed
*   error, if fail
 */
func (fa *FileAgent) FileRecheck() error {
	filename := fa.Filename

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("file %s opening FAIL: %s", fa.Filename, err.Error())
		fa.UnchangeTime = 0
		return err
	}

	fileinfo, err := file.Stat()
	if err != nil {
		log.Printf("file %s stat FAIL: %s", fa.Filename, err.Error())
		fa.UnchangeTime = 0
		return err
	}

	isSameFile := os.SameFile(fa.FileInfo, fileinfo)
	if !isSameFile {
		log.Printf("file %s recheck, it is a new file", fa.Filename)
		if fa.File != nil {
			if err := fa.File.Close(); err != nil {
				log.Printf("old file closing FAIL: %s", err.Error())
			}
		}

		fa.File = file
		fa.FileInfo = fileinfo
		fa.LastOffset = 0
		fa.UnchangeTime = 0

		// seek the cursor to the end of new file
		offset, err := fa.File.Seek(fa.FileInfo.Size(), os.SEEK_SET)
		if err != nil {
			log.Printf("seek file %s FAIL: %s", fa.Filename, err.Error())
		}
		log.Printf("seek file %s to %d", fa.Filename, offset)
		fa.LastOffset += fa.FileInfo.Size()

		// initialize agent data
		now := time.Now()
		minute := now.Format("200601021504")

		tsNow := now.Unix()
		tsStart := tsNow

		start, err := time.ParseInLocation("20060102150405", minute+"00", now.Location())
		if err != nil {
			log.Printf("timestamp setting FAIL: %s", err.Error())
		} else {
			tsStart = start.Unix()
		}

		fa.Data.TsStart = tsStart
		fa.Data.TsEnd = tsStart + fa.Data.Step - 1
		fa.Data.TsUpdate = tsNow
		fa.Data.ErrorCnt = 0
		fa.Data.ReqCnt = 0
		fa.Data.Req200Cnt = 0
		fa.Data.Req499Cnt = 0
		fa.Data.Req500Cnt = 0
		fa.Data.Req502Cnt = 0
		fa.Data.Req504Cnt = 0
		fa.Data.IpSet = NewStringSet()
		fa.Data.IdSet = NewStringSet()

		return nil
	} else {
		fa.UnchangeTime = 0
		return nil
	}
}

/*
* IsChanged - check the change of log file
*
* RECEIVER: *FileAgent
*
* PARAMS:
*   No paramter
*
* RETURNS:
*   - true: if change
*   - false: if not change
 */
func (fa *FileAgent) IsChanged() bool {
	lastMode := fa.FileInfo.Mode()
	lastSize := fa.FileInfo.Size()
	lastModTime := fa.FileInfo.ModTime().Unix()

	fileinfo, err := fa.File.Stat()
	if err != nil {
		log.Printf("file %s stat FAIL: %v", err)
		fa.UnchangeTime += 1
		return false
	}

	thisMode := fileinfo.Mode()
	thisSize := fileinfo.Size()
	thisModTime := fileinfo.ModTime().Unix()
	thisTailable := fileinfo.Mode().IsRegular()

	if lastMode == thisMode &&
		(!thisTailable || lastSize == thisSize) &&
		lastModTime == thisModTime {

		fa.UnchangeTime += 1
		return false
	}

	// replace the FileInfo for reading the new content
	fa.UnchangeTime = 0
	fa.FileInfo = fileinfo
	return true
}
