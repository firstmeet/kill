package pkg

import "C"
import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
)

var WhiteListName = []string{"telnet", "sh", "sudo", "ssh"}

type Element interface {
	string | int | int32 | int64
}

func GetNeedKillPids(duration int64) ([]int32, []string) {
	var result = make([]int32, 0)
	var names = make([]string, 0)
	pids, err := getPids()
	if err != nil {
		return nil, nil
	}
	if len(pids) <= 0 {
		return nil, nil
	}
	systemStartTime := GetSystemStartTime()
	if systemStartTime == 0 {
		return nil, nil
	}
	for _, p1 := range pids {
		processNew, err := process.NewProcess(p1)
		if err != nil {
			continue
		}
		pocessCreateTime, err := processNew.CreateTime()
		if err != nil {
			continue
		}
		if pocessCreateTime/1000-systemStartTime >= duration && int32(SelfPid()) != p1 {
			name, err := processNew.Name()
			if err != nil {
				continue
			}
			if SliceContain(WhiteListName, name) {
				continue
			} else {
				result = append(result, p1)
				names = append(names, name)
			}
		}
	}
	return result, names
}
func getPids() ([]int32, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}
	return pids, nil
}
func SelfPid() int {
	pid := os.Getpid()
	return pid
}
func KillPids(pids []int32) {
	myPid := os.Getpid()
	getppid := os.Getppid()
	for _, pid := range pids {
		if pid != int32(myPid) && pid != int32(getppid) {
			newProcess, err := process.NewProcess(pid)
			if err != nil {
				continue
			}
			err = newProcess.Kill()
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}
func KillPid(process *process.Process) {
	err := process.Kill()
	if err != nil {
		fmt.Println(err)
		return
	}
}
func GetSystemStartTime() int64 {
	command := exec.Command("cat", "/proc/uptime")
	output, err := command.Output()
	if err != nil {
		return 0
	}
	split := strings.Split(string(output), " ")
	timeFloat := split[0]
	seconds := strings.Split(timeFloat, ".")
	secondsInt, err := strconv.Atoi(seconds[0])
	i := int64(secondsInt)
	startTime := time.Now().Unix() - i
	return startTime
}
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
