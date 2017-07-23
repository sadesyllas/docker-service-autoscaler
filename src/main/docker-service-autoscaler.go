package main

import (
	"bytes"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"../cluster"
	"../service"
)

func init() {

}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	configNotProvided := len(os.Args) < 3

	if !configNotProvided {
		if stat, err := os.Stat(os.Args[1]); os.IsPermission(err) || os.IsNotExist(err) || stat.IsDir() {
			configNotProvided = true
		}
	}

	/*if configNotProvided { // TODO: uncomment
		log.Fatal("usage: docker-service-autoscaler \"/path/to/json/config/file\" \"/path/to/log/file\"\n")
	}

	if os.Args[2] != "-" {
		if stat, err := os.Stat(os.Args[2]); os.IsPermission(err) || stat.IsDir() {
			log.Fatalf("cannot log to directory %s", os.Args[2])
		}

		logFile, err := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE, 0755)

		if err != nil {
			log.Fatalf("cannot log to file %s: %s", os.Args[2], err)
		}

		defer logFile.Close()

		log.SetOutput(logFile)
	}*/

	for {
		mainRecovering( /*os.Args[1]*/ "") // TODO: uncomment
	}
}

func mainRecovering(configPath string) {
	sigHUP := make(chan os.Signal, 1)
	signal.Notify(sigHUP, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL)

	defer func() {
		if err := recover(); err != nil {
			var buf bytes.Buffer
			s := err.(string)
			buf.WriteString(s)
			buf.WriteString("\n")
			os.Stderr.WriteString(buf.String())
		}

		signal.Reset(syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL)
		close(sigHUP)
	}()

	schedule(func() { service.UpdateConfig(configPath) }, "config update")
	schedule(func() { cluster.UpdateState() }, "cluster state")
	schedule(func() { service.ScaleServices() }, "services scaling")

	<-sigHUP
	os.Exit(0)
}

func schedule(f func(), desc string) {
	log.Debugf("scheduling %s", desc)

	f()

	time.Sleep(time.Duration(5) * time.Second)

	go schedule(f, desc)
}
