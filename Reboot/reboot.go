package reboot

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"time"
)

func Reboot(addr string, id string) {
	count := backupPhase()
	go primaryPhase(count, addr, id)
}

func backupPhase() int {
	fmt.Println("--- Backup phase ---")
	countOld, err := ioutil.ReadFile("Reboot/count.data")
	if err != nil {
		fmt.Println("... timed out")
		return 0
	}
	for {
		time.Sleep(2*time.Second)
		countNew, err := ioutil.ReadFile("Reboot/count.data")
		if err != nil {
			fmt.Println(err)
		}
		if string(countNew) == string(countOld) {
			fmt.Println("... timed out")
			count, _ := strconv.Atoi(string(countNew))
			return count
		}
		countOld = countNew
	}
}

func primaryPhase(count int, addr string, id string) {
	fmt.Println("--- Primary phase ---")
	fmt.Println("... creating new backup")
	err := exec.Command("cmd", "/C", "start", "powershell", "go", "run", "main.go", "-addr="+addr, "-id="+id).Run()
	if err != nil {
		fmt.Println(err)
	}
	for {
		count++
		err := ioutil.WriteFile("Reboot/count.data", []byte(strconv.Itoa(count)), 0777)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second)
	}
}
