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
	"time"
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

// DoTask is Hdd test
func DoTask(label int, tc TestConfig) {
	defer func() {
		locks[label] = false
	}()
	task := &tasks{
		// mu:        &sync.Mutex{},
		// logfile:   &os.File{},
		// tasklist:  []string{},
		lastError: nil,
		label:     label,
		// cmddict:   map[int]*exec.Cmd{},
		linuxName: "",
		sgName:    "",
	}
	err := task.Init(label)
	defer task.Uninit()
	if err != nil {
		return
	}
	for _, tt := range tc.Items {
		task.logIt("start task: " + tt.Name)
	}
}

func (t *tasks) logIt(s string) {
	t.logfile.WriteString(s)
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
	t.logfile.WriteString(fmt.Sprintf("task start time: %s\n", time.Now().Format("2006-01-02 15:04:05")))
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
	t.logfile.WriteString(fmt.Sprintf("task end time: %s\n", time.Now().Format("2006-01-02 15:04:05")))
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

func (t *tasks) runExe(title string, parser func(line string) error, name string, args ...string) error {
	if title != "" {
		t.logfile.WriteString(fmt.Sprintf("\n\n...........................:  %s  :...........................\n", title))
	}
	cmd := exec.Command(name, args...)
	// Get a pipe to read from standard out
	r, _ := cmd.StdoutPipe()
	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout
	// Make a new channel which will be used to ensure we get all output
	done := make(chan bool)

	// Create a scanner which scans r in a line-by-line fashion
	scanner := bufio.NewScanner(r)
	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {

		// Read line by line and process it
		for scanner.Scan() {
			line := scanner.Text()
			//HandleLog(label, line)
			t.logfile.WriteString(line + "\n")
			parser(line)
		}
		// We're all done, unblock the channel
		done <- true

	}()
	// Start the command and check for errors
	err := cmd.Start()
	t.cmddict[cmd.Process.Pid] = cmd
	// Wait for the command to finish
	err = cmd.Wait()
	// Wait for all output to be processed
	<-done
	delete(t.cmddict, cmd.Process.Pid)
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

func (t *tasks) ReadInquiry() error {

	Parser := func(sline string) {
		if strings.HasPrefix(sline, "[") {
			return
		}
		slc := strings.Split(sline, ":")
		for i := range slc {
			slc[i] = strings.TrimSpace(slc[i])
		}
		if len(slc) == 2 && slc[0] != "" && slc[1] != "" {
			Set(t.label, slc[0], slc[1], 0)
		}
	}

	t.logfile.WriteString("\n\n...........................:  Driver Inquiry Data  :...........................\n")
	cmd := exec.Command("./sg_inq", t.sgName)
	// Get a pipe to read from standard out
	r, _ := cmd.StdoutPipe()
	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout
	// Make a new channel which will be used to ensure we get all output
	done := make(chan bool)

	// Create a scanner which scans r in a line-by-line fashion
	scanner := bufio.NewScanner(r)
	// scanner.Split(ScanItems)
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
	delete(t.cmddict, cmd.Process.Pid)
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

func (t *tasks) DriverIdentifyData() error {
	// hdparm -I /dev/sdd     //sg_readcap -l /dev/sg4
	var supported bool
	Parser := func(sline string) {
		s := strings.Trim(sline, " \t")

		if s == "Enabled\tSupported:" {
			supported = true
			t.logfile.WriteString("find supported List:\n")
		}
		if supported {
			t.logfile.WriteString(s + "\n")
		}
	}

	t.logfile.WriteString("\n\n...........................:  Driver Identify Data  :...........................\n")
	cmd := exec.Command("hdparm", "-I", t.sgName)
	// Get a pipe to read from standard out
	r, _ := cmd.StdoutPipe()
	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout
	// Make a new channel which will be used to ensure we get all output
	done := make(chan bool)

	// Create a scanner which scans r in a line-by-line fashion
	scanner := bufio.NewScanner(r)
	// scanner.Split(ScanItems)
	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {

		// Read line by line and process it
		for scanner.Scan() {
			line := scanner.Text()

			// t.logfile.WriteString(line + "\n")
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
	delete(t.cmddict, cmd.Process.Pid)
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

func (t *tasks) TestIDCheck() error {
	// sg_vpd -a /dev/sg4
	Parser := func(line string) error {
		return nil
	}
	return t.runExe("ID Check", Parser, "./sg_vpd", "-a", t.sgName)
}

func (t *tasks) TestMaxAddress() error {
	// hdparm -N
	var maxaddr string
	var bSame bool
	Parser := func(line string) error {
		bSame = true
		return nil
	}
	err := t.runExe("Set Native Max Address", Parser, "hdparm", "-N", t.linuxName)
	if err != nil {
		return err
	}
	if bSame {
		t.logfile.WriteString("Native Max Address is same, No change")
		return nil
	}
	return t.runExe("", Parser, "hdparm", "-N", maxaddr, t.linuxName)
}

func (t *tasks) TestCryptoScramble() error {
	var errContinue error
	Parser := func(line string) error {
		return nil
	}
	err := t.runExe("Crypto Scramble", Parser, "hdparm", "--yes-i-know-what-i-am-doing", "--sanitize-crypto-scramble", t.linuxName)
	if err != nil {
		return err
	}
	for {
		err := t.runExe("", Parser, "hdparm", "--sanitize-status", t.linuxName)
		if err == errContinue {
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func (t *tasks) TestSmartCheck() error {
	Parser := func(line string) error {
		return nil
	}
	err := t.runExe("Smart Test", Parser, "smartctl", "-a", t.linuxName)
	if err != nil {
		return err
	}
	return nil
}

func (t *tasks) TestModeSense() error {
	Parser := func(line string) error {
		return nil
	}
	err := t.runExe("Mode Sense Test", Parser, "./sginfo", "-a", t.sgName)
	if err != nil {
		return err
	}
	return nil
}

func (t *tasks) TestGListCheck() error {
	//get data badsector
	return nil
}

func (t *tasks) TestFillData(dd byte) error {
	Parser := func(line string) error {
		return nil
	}
	err := t.runExe("Fill Data", Parser, "./dskwipe", "-y", "-n", "8000", fmt.Sprintf("0x%2X", dd), t.sgName)
	if err != nil {
		return err
	}
	return nil
}

func (t *tasks) TestButterfly() error {
	Parser := func(line string) error {
		return nil
	}
	err := t.runExe("Butterfly Test", Parser, "./openSeaChest_GenericTests", "--butterflyTest", "--minutes", "5", "-d", t.sgName)
	if err != nil {
		return err
	}
	return nil
}

func (t *tasks) TestRandom() error {
	Parser := func(line string) error {
		return nil
	}
	err := t.runExe("Random Blank Test", Parser, "./openSeaChest_GenericTests", "--randomTest", "-d", t.sgName)
	if err != nil {
		return err
	}
	return nil
}

func (t *tasks) TestVerifyData(dd byte) error {
	Parser := func(line string) error {
		return nil
	}
	err := t.runExe("Verify Data", Parser, "./dskread", "-p", fmt.Sprintf("0x%2X", dd), t.sgName)
	if err != nil {
		return err
	}
	return nil
}
