package main

import (
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/process"
	"kill/pkg"
	"os"
	"strings"
	"time"
)

func main() {
	duration := flag.Int64("dd", 100, "")
	whiteList := flag.String("w", "telnet,ssh,sh,wget,curl,sudo", "")
	pkg.WhiteListName = strings.Split(*whiteList, ",")
	flag.Parse()
	pids, names := pkg.GetNeedKillPids(*duration)
	fmt.Println(names)
	go killPids(pids)
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ticker.C:
				pids, _ := pkg.GetNeedKillPids(*duration)
				go killPids(pids)
			}
		}
	}()
	select {}
}
func killPids(pids []int32) {
	myPid := os.Getpid()
	getppid := os.Getppid()
	fmt.Println("myPid:", myPid)
	fmt.Println("ppid:", getppid)
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
