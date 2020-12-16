package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
)

type tasks struct {
	mu        *sync.Mutex
	logfile   *os.File
	tasklist  []string
	lastError error
	label     int
	cmddict   map[int]*exec.Cmd
	linuxName string
	sgName    string
}

// Init create Log File
func (t *tasks) Init(ll int) error {
	t.label = ll
	folder := path.Join(os.Getenv("DSEDHOME"), "logs", fmt.Sprintf("label_%d", ll))
	os.MkdirAll(folder, os.ModePerm)

	logs, err := os.OpenFile(fmt.Sprintf("%s/log_%d.log", folder, t.label), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		Set(ll, "errorcode", 13, 0)
		return err
	}
	t.logfile = logs
	t.mu = &sync.Mutex{}
	t.cmddict = make(map[int]*exec.Cmd)

	sdevname, err := GetString(ll, "linuxname")
	if err != nil {
		fmt.Println("linuxname not found")
		Set(ll, "errorcode", 11, 0)
		return err
	}
	t.linuxName = sdevname
	if len(sdevname) > 0 {
		exec.Command("umount", sdevname).Output()
	}
	sgName, err := GetString(ll, "sglibName")
	if err != nil {
		fmt.Println("sglibName not found")
		Set(ll, "errorcode", 12, 0)
		return err
	}
	t.sgName = sgName
	return nil
}

// Uninit Exit
func (t *tasks) Uninit() {
	t.logfile.Close()
}

// GetTaskList for single port to do
func (t *tasks) GetTaskList() error {
	cnt, err := GetListsCnt("tasklist")
	if err != nil {
		return err
	}
	t.tasklist, err = GetLists("tasklist", 0, cnt)
	if err != nil {
		return err
	}
	return nil
}

func (t *tasks) ReadInquiry() error {

	Parser := func(sline string) {
		slc := strings.Split(sline, ":")
		for i := range slc {
			slc[i] = strings.TrimSpace(slc[i])
		}
		if len(slc) == 2 && slc[0] != "" && slc[1] != "" {
			Set(t.label, slc[0], slc[1], 0)
		}
	}

	t.logfile.WriteString("...........................:  Driver Inquiry Data  :...........................")
	cmd := exec.Command("./sg_inq", t.sgName)
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
			t.logfile.WriteString(line + "\n")
			Parser(line)
		}
		// We're all done, unblock the channel
		done <- true

	}()
	// Start the command and check for errors
	err := cmd.Start()
	t.cmddict[cmd.Process.Pid] = cmd
	// Wait for all output to be processed
	<-done
	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			t.logfile.WriteString(fmt.Sprintf("ExitCode=%d\n", waitStatus.ExitStatus()))
		}
		return err
	}
	// Success
	waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
	t.logfile.WriteString(fmt.Sprintf("ExitCode=%d\n", waitStatus.ExitStatus()))

	return nil
}