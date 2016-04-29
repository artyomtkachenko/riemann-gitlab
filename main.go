package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/amir/raidman"
)

type serviceStatus struct {
	status   string
	duration int //seconds
}

func logErr(err error) {
	if err != nil {
		log.Printf("%+v", err)
	}
}

func parseGitlabStatus(input []byte) map[string]serviceStatus {
	var result = make(map[string]serviceStatus)
	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		if strings.Contains(line, "log:") {
			s := strings.Split(line, " ")
			service := s[1][:len(s[1])-1]

			if s[0] == "run:" {
				result[service] = serviceStatus{"ok", getDuration(s[4])}
			} else if s[0] == "down:" {
				result[service] = serviceStatus{"critical", getDuration(s[2])}
			}
		}
	}
	return result
}

func getDuration(s string) int {
	duration, err := strconv.Atoi(s[:len(s)-2])
	logErr(err)

	return duration
}

func main() {
	host := flag.String("host", "", "Riemann Collector hostname")
	port := flag.String("port", "", "Riemann Collector port")
	services := flag.String("service", "gitlab-workhorse,logrotate,nginx,sidekiq,unicorn,redis", "Comma-separated gitlab serivices to look after")
	upDown := map[string]string{
		"ok":       "up",
		"critical": "down",
	}
	logwriter, e := syslog.New(syslog.LOG_NOTICE, "riemann-gitlab")
	if e == nil {
		log.SetOutput(logwriter)
	}

	c, err := raidman.DialWithTimeout("tcp", fmt.Sprintf("%s:%s", *host, *port), time.Duration(10)*time.Second)
	logErr(err)

	servicesToCheck := strings.Split(*services, ",")

	for {
		output, err := exec.Command("bash", "-c", "gitlab-ctl status").Output()
		logErr(err)

		result := parseGitlabStatus(output)

		for _, item := range servicesToCheck {
			err := c.Send(&raidman.Event{
				Service:     "gitab-" + item,
				State:       result[item].status,
				Metric:      result[item].duration,
				Description: fmt.Sprintf("GitLab's %s has been %s for %d seconds", item, upDown[result[item].status], result[item].duration),
				Tags:        []string{"gitlab", item},
			})
			logErr(err)
		}
		log.Print("Tick")
		time.Sleep(10 * time.Second)
	}
	defer c.Close()
	defer logwriter.Close()
}
