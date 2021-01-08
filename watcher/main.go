package main

import (
	"flag"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	flag.Parse()
	var ips []string
	if len(flag.Args()) != 0 {
		ips = flag.Args()
	} else {
		ips = []string{"192.168.2.1"}
	}
	log.Printf("Check %s\n", ips)
	receivedCount := 0
	for _, ip := range ips {
		command := exec.Command("ping", "-c", "5", ip)
		output, _ := command.Output()
		s := string(output)
		split := strings.Split(s, "\n")
		for _, t := range split {
			if strings.Contains(t, "received") && strings.Contains(t, "loss") {
				log.Println(t)
				i := strings.Split(strings.TrimSpace(strings.Split(t, ",")[1]), " ")
				count, _ := strconv.Atoi(strings.TrimSpace(i[0]))
				receivedCount += count
				break
			}
		}
	}
	if receivedCount == 0 {
		log.Println("It is Dead")
		//crontab should use absolute path
		err := exec.Command("/usr/sbin/shutdown", "now").Start()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("It is Alive")
	}
}
