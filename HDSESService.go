// HDSESService
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"regexp"

	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var (
	msgOK    = []byte(`{"result":"ok"}`)
	msgError = []byte(`{"result":"error"}`)
)

type writer struct {
	mu *sync.Mutex
	wl *os.File
}

func (w *writer) Write(bytes []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	//fmt.Printf("%s\n", string(bytes))
	w.wl.WriteString(string(bytes) + "\n")
	return len(bytes), nil
}

type processlabel struct {
	cmddict map[int]*exec.Cmd
	mu      *sync.Mutex
}

func (pl *processlabel) Add(label int, cmd *exec.Cmd) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if _, ok := pl.cmddict[label]; ok {
		delete(pl.cmddict, label)
	}
	//fmt.Printf("Add: %d  %v\n", label, cmd)
	pl.cmddict[label] = cmd
}

func (pl *processlabel) Remove(label int) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if cmd, ok := pl.cmddict[label]; ok && cmd != nil {
		fmt.Printf("Remove: %d ready kill\n", label)
		cmd.Process.Kill()
		delete(pl.cmddict, label)
	}
}

// IsSSD check is SSD
func IsSSD(devicename string) bool {
	isSSD := false
	ss, _ := exec.Command("smartctl", "-i", devicename).Output()
	if strings.Contains(string(ss), "Solid State Device") {
		isSSD = true
	}
	return isSSD
}

func divmod(numerator, denominator int64) (quotient, remainder int64) {
	quotient = numerator / denominator // integer division, decimals are truncated
	remainder = numerator % denominator
	return
}

// ScanItems 逗号或分号 的自定义分隔
func ScanItems(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r"); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

// RunExeWipe run dskwipe and handle output to database
func RunExeWipe(logpath string, devicename string, patten string, label int) error {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dskwipe := path.Join(dir, "dskwipe")
	fmt.Printf("%s %s %s %s %s %s\n", dskwipe, devicename, "-y", "-n", "8000", patten)
	Set(label, "starttasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
	SetTransaction(label, "StartTime", time.Now().Format("2006-01-02 15:04:05Z"))
	cmd := exec.Command(dskwipe, devicename, "-y", "-n", "8000", patten)

	processlist.Add(label, cmd)

	f, err := os.OpenFile(fmt.Sprintf("%s/log_%d.log", logpath, label), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		Set(label, "errorcode", 1, 0)
		SetTransaction(label, "errorCode", 100)
		Set(label, "endtasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
		// Publish(label, "taskdone", 1)
		PublishTaskDone(label, 20)
		return err
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%s %s %s %s %s %s\n", dskwipe, devicename, "-y", "-n", "8000", patten))

	// Get a pipe to read from standard out
	r, _ := cmd.StdoutPipe()

	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout

	// Make a new channel which will be used to ensure we get all output
	done := make(chan bool)

	// Create a scanner which scans r in a line-by-line fashion
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanItems)
	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {

		// Read line by line and process it
		for scanner.Scan() {
			line := scanner.Text()
			//HandleLog(label, line)
			f.WriteString(line + "\n")
			handlelogprogress(label, line)
		}

		// We're all done, unblock the channel
		done <- true

	}()

	// Start the command and check for errors
	err = cmd.Start()

	// Wait for all output to be processed
	<-done

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			fmt.Printf("WipeExitCode=%d\n", waitStatus.ExitStatus())
			f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", waitStatus.ExitStatus()))
			Set(label, "errorcode", waitStatus.ExitStatus(), 0)
			SetTransaction(label, "errorCode", waitStatus.ExitStatus())
		}
	} else {
		// Success
		waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
		fmt.Printf("WipeExitCode=%d\n", waitStatus.ExitStatus())
		f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", waitStatus.ExitStatus()))
		Set(label, "errorcode", waitStatus.ExitStatus(), 0)
		SetTransaction(label, "errorCode", waitStatus.ExitStatus())
	}
	Set(label, "endtasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
	// Publish(label, "taskdone", 1)
	PublishTaskDone(label, 21)
	return err
}

// RunSecureErase Run Secure Erase
func RunSecureErase(logpath string, devicename string, label int) {
	f, err := os.OpenFile(fmt.Sprintf("%s/log_%d.log", logpath, label), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		Set(label, "errorcode", 1, 0)
		SetTransaction(label, "errorCode", 100)
		Set(label, "endtasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
		// Publish(label, "taskdone", 1)
		PublishTaskDone(label, 20)
		return
	}
	defer f.Close()
	tstart := time.Now()
	f.WriteString(fmt.Sprintf("Start Task local time and date: %s\n", tstart.Format("Mon Jan _2 15:04:05 2006")))
	Set(label, "starttasktime", tstart.Format("Mon Jan _2 15:04:05 2006"), 0)
	values := []string{"1", "1", "0x00", "0.00%", "0.00%", "00:01", "", "", "00:01", "0.00", "0.00"}
	setProgressbar(label, values)
	SetTransaction(label, "StartTime", time.Now().Format("2006-01-02 15:04:05Z"))
	stime := tstart.Format("15:04:05")
	funReadData := func() (string, error) {
		// if sector size is 520, this code is not working.Must use sglib. but not find go sglib.
		f, err := syscall.Open(devicename, syscall.O_RDONLY, 0777)
		if err != nil {
			log.Fatal(err)
			return "", err
		}
		defer syscall.Close(f)
		b1 := make([]byte, 512)
		_, err = syscall.Read(f, b1)
		if err != nil {
			return "", err
		}
		md5 := md5.Sum(b1)
		ss := fmt.Sprintf("%x", md5)
		return ss, nil
	}

	funWriteData := func() error {
		// if sector size is 520, this code is not working.Must use sglib. but not find go sglib.
		f, err := syscall.Open(devicename, syscall.O_WRONLY, 0777)
		if err != nil {
			log.Fatal(err)
			return err
		}
		defer syscall.Close(f)
		b1 := make([]byte, 512)
		for i := 0; i < 512; i++ {
			b1[i] = 65
		}
		_, err = syscall.Write(f, b1)
		if err != nil {
			return err
		}
		return nil
	}

	var errorcode int
	smd5, err := funReadData()
	if err != nil {
		errorcode = 10
	}

	bverify := false

	if IsSSD(devicename) {
		funWriteData()
		f.WriteString(fmt.Sprintf("hdparm --user-master u --security-set-pass PASSFD %s\n", devicename))
		exec.Command("hdparm", "--user-master", "u", "--security-set-pass", "PASSFD", devicename).Output()
		f.WriteString(fmt.Sprintf("hdparm --user-master u --security-erase PASSFD %s\n", devicename))
		exec.Command("hdparm", "--user-master", "u", "--security-erase", "PASSFD", devicename).Output()
		tend := (int64)(time.Now().Sub(tstart).Seconds())
		hours, remainder := divmod(tend, 3600)
		minutes, seconds := divmod(remainder, 60)
		send := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		//   1      1 0x00   6.737%   6.737% 00:02:30 00:02:30 17:00:51 00002226   134.73   134.73
		line := fmt.Sprintf("   1      1 0x00 100.000%% 100.000%% %s %s %s %08d     0.00     0.00\n", send, send, stime, tend)
		f.WriteString(line)
		handlelogprogress(label, line)
		f.WriteString(fmt.Sprintf("end Task local time and date: %s\n", time.Now().Format("Mon Jan _2 15:04:05 2006")))
		//f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", 0))
		// Set(label, "endtasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
		smd51, err := funReadData()
		if err != nil {
			errorcode = 10
		}
		if errorcode == 0 {
			bverify = smd51 == smd5
		}
		values := []string{"1", "1", "0x00", "100.00%", "100.00%", send, send, stime, fmt.Sprintf("%d", tend), "0.00", "0.00"}
		setProgressbar(label, values)

	} else {
		f.WriteString(fmt.Sprintf("hdparm --yes-i-know-what-i-am-doing --sanitize-crypto-scramble %s\n", devicename))
		data, err := exec.Command("hdparm", "--yes-i-know-what-i-am-doing", "--sanitize-crypto-scramble", devicename).CombinedOutput()
		if err != nil {
			errorcode = 100
		}
		f.WriteString(string(data))
		if strings.IndexAny(string(data), "is not supported") < 0 {
			time.Sleep(2 * time.Second)
			f.WriteString(fmt.Sprintf("hdparm --sanitize-status %s\n", devicename))
			exec.Command("hdparm", "--sanitize-status", devicename).Output()
			time.Sleep(2 * time.Second)
			exec.Command("hdparm", "--sanitize-status", devicename).Output()
		} else {
			errorcode = 100 //not support
			f.WriteString(fmt.Sprintf("error=%v\n", err))
		}
		tend := (int64)(time.Now().Sub(tstart).Seconds())
		hours, remainder := divmod(tend, 3600)
		minutes, seconds := divmod(remainder, 60)
		send := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		//   1      1 0x00   6.737%   6.737% 00:02:30 00:02:30 17:00:51 00002226   134.73   134.73
		line := fmt.Sprintf("   1      1 0x00 100.000%% 100.000%% %s %s %s %08d     0.00     0.00\n", send, send, stime, tend)
		f.WriteString(line)
		handlelogprogress(label, line)
		f.WriteString(fmt.Sprintf("end Task local time and date: %s\n", time.Now().Format("Mon Jan _2 15:04:05 2006")))
		//f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", 0))
		// Set(label, "endtasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
		if errorcode != 100 {
			smd51, err := funReadData()
			if err != nil {
				errorcode = 10
			}
			if errorcode == 0 {
				bverify = smd51 != smd5
			}
		}
		values := []string{"1", "1", "0x00", "100.00%", "100.00%", "00:01", send, stime, "00:01", "0.00", "0.00"}
		setProgressbar(label, values)
	}

	if bverify && errorcode == 0 {
		f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", 0))
		Set(label, "errorcode", 0, 0)
		SetTransaction(label, "errorCode", 0)
	} else {
		f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", errorcode))
		Set(label, "errorcode", errorcode, 0)
		SetTransaction(label, "errorCode", errorcode)
	}
	Set(label, "endtasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
	// Publish(label, "taskdone", 1)
	PublishTaskDone(label, 23)
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02d:%02d", h, m)
}

func handlelogprogress(label int, line string) {
	var validlog = regexp.MustCompile(`(\d*\.\d*)%.*?(\d*\.\d*)%.*?(\d*\.\d*)$`)
	if !validlog.MatchString(line) {
		fmt.Println("Not Match:" + line)
		return
	}
	sp := func(r rune) bool {
		return r == '\t' || r == ' '
	}
	infos := strings.FieldsFunc(line, sp)
	//write database
	infos[9] = strings.Split(infos[9], ".")[0] + " MB/s"
	infos[5] = infos[5][:len(infos[5])-3]

	secondtotime := func(s string) string {
		ret := "00:01"
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return ret
		}
		ti := time.Duration(i) * time.Second
		ret = fmtDuration(ti)
		return ret
	}
	infos[8] = secondtotime(infos[8])

	goprogress := func(s string) string {
		ret := "0.01%"
		i, err := strconv.ParseFloat(s[:len(s)-1], 64)
		if err != nil {
			return ret
		}
		if i < 1.0 {
			i = 1.0
		} else if i > 99.0 {
			i = 99.0
		}
		ret = fmt.Sprintf("%.02f%%", i)
		return ret
	}
	infos[4] = goprogress(infos[4])

	fmt.Println(infos)
	if err := setProgressbar(label, infos); err != nil {
		//print log
		fmt.Println(err)
	}

}

// RunWipe call dskwipe
func RunWipe(logpath string, devicename string, patten string, label int) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dskwipe := path.Join(dir, "dskwipe")
	fmt.Printf("%s %s %s %s %s %s\n", dskwipe, devicename, "-y", "-n", "8000", patten)
	Set(label, "starttasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
	SetTransaction(label, "StartTime", time.Now().Format("2006-01-02 15:04:05Z"))
	cmd := exec.Command(dskwipe, devicename, "-y", "-n", "8000", patten)

	processlist.Add(label, cmd)

	f, err := os.OpenFile(fmt.Sprintf("%s/logs/%s/log_%d.log", os.Getenv("DSEDHOME"), logpath, label), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%s %s %s %s %s %s\n", dskwipe, devicename, "-y", "-n", "8000", patten))

	var mu sync.Mutex

	cmd.Stderr = &writer{
		mu: &mu,
		wl: f,
	}
	cmd.Stdout = &writer{
		mu: &mu,
		wl: f,
	}
	/*
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		err = cmd.Wait()
		if err != nil {
			log.Fatal(err)
		}
	*/
	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
			f.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			fmt.Printf("WipeExitCode=%d\n", waitStatus.ExitStatus())
			f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", waitStatus.ExitStatus()))
			Set(label, "errorcode", waitStatus.ExitStatus(), 0)
			SetTransaction(label, "errorCode", waitStatus.ExitStatus())
		}
	} else {
		// Success
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		fmt.Printf("WipeExitCode=%d\n", waitStatus.ExitStatus())
		f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", waitStatus.ExitStatus()))
		Set(label, "errorcode", waitStatus.ExitStatus(), 0)
		SetTransaction(label, "errorCode", waitStatus.ExitStatus())
	}
	Set(label, "endtasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
}

// var mu sync.Mutex
var processlist *processlabel

var configxmldata *configs

func main() {
	fmt.Println("hdsesserver version: 20.12.15.0, auther:Jeffery Zhang")
	runtime.GOMAXPROCS(4)

	processlist = &processlabel{
		cmddict: make(map[int]*exec.Cmd),
		mu:      &sync.Mutex{},
	}

	LoadConfigXML()

	r := mux.NewRouter()
	// Add your routes as needed
	r.HandleFunc("/start/{label:[0-9]+}", startTaskHandler).Methods("GET").Queries("standard", "{standard}")
	r.HandleFunc("/stop/{label:[0-9]+}", stopTaskHandler).Methods("GET")

	CreateRedisPool(GetLabelsCnt())
	srv := &http.Server{
		Addr: "0.0.0.0:12100",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)

}
