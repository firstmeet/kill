package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"

	"github.com/vishvananda/netlink"
)

type Element interface {
	string | int | int32 | int64
}

var WhiteListName = []string{
	"telnet", "ps", "top", "sh", "sudo", "ssh", "zsh", "bash", "ash",
	"cron", "systemd", "systemd-journald", "dbus-daemon",
	"systemd-logind", "systemd-timesyncd", "accounts-daemon",
	"ksmd", "kcompactd", "khungtaskd",
	"crypto", "kintegrityd", "ext4", "kblockd", "kswapd", "/usr/bin", "/usr/sbin",
	"gdbus", "gmain", "worker",
	"systemd-resolv"}

func SliceContain[T Element](arr []T, target T) bool {
	for _, t := range arr {
		if reflect.TypeOf(target).Kind().String() == "string" {
			if strings.Contains(string(target), string(t)) {
				return true
			}
		} else {
			if t == target {
				return true
			}
		}

	}
	return false
}
func tProcEvent(duration int64) {
	ch := make(chan netlink.ProcEvent)
	done := make(chan struct{})
	defer close(done)

	errChan := make(chan error)

	if err := netlink.ProcEventMonitor(ch, done, errChan); err != nil {
		fmt.Printf("err: %v\n", err)
	}
	for {
		e := <-ch
		if e.Msg.Tgid() == uint32(os.Getpid()) {
			continue
		}
		newProcess, err := process.NewProcess(int32(e.Msg.Pid()))
		if err != nil {
			continue
		}
		name, _ := newProcess.Name()

		startTime := getSystemStartTime()
		if startTime == 0 {
			continue
		}
		processTime, err2 := newProcess.CreateTime()
		if err2 != nil {
			continue
		}
		if processTime/1000-startTime < duration {
			continue
		}

		if SliceContain(WhiteListName, name) {
			continue
		}
		fmt.Printf("e.Msg: %v,%#v\n", e.Msg.Pid(), name)
		KillPid(newProcess)
	}

	done <- struct{}{}
}
func getSystemStartTime() int64 {
	command := exec.Command("cat", "/proc/uptime")
	output, err := command.Output()
	if err != nil {
		return 0
	}
	split := strings.Split(string(output), " ")
	timeFloat := split[0]
	seconds := strings.Split(timeFloat, ".")
	secondsInt, err := strconv.Atoi(seconds[0])
	if err != nil {
		return 0
	}
	i := int64(secondsInt)
	startTime := time.Now().Unix() - i
	return startTime
}
func KillPid(process *process.Process) {
	err := process.Kill()
	if err != nil {
		fmt.Println(err)
		return
	}
	// s, _ := process.Name()
	// fmt.Println("kill pid:", process.Pid, s)
}
func getProcByPid(pid uint32) (out []byte) {

	t, _ := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	for i := 0; i < len(t); i++ {
		if t[i] == 0 {
			t[i] = 0x20
		}
	}
	return t
}
func main() {

	tProcEvent(300)
}
