package daemon

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/srevinsaju/zap/internal/helpers"
	"github.com/srevinsaju/zap/logging"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var logger = logging.GetLogger()

type UpdateFunction func() ([]string, error)

func upgrade(c chan int, s chan os.Signal, updater UpdateFunction) {

	for {
		select {
		case <-c:
			apps, _ := updater()
			if len(apps) > 0 {
				logger.Infof("Apps have been updated, %s", apps)
				err := beeep.Notify("Zap ⚡️",
					fmt.Sprintf("Updated %s", strings.Join(apps, ", ")), "")
				if err != nil {
					panic(err)
				}
			} else {
				logger.Infof("All apps up-to-date")
			}

		case <-s:
			fmt.Println("quit")
			return
		}
	}
}

func Sync(updater UpdateFunction) {

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)

	c := make(chan int)
	count := 0

	go func() {

		for {
			waitUntilOnline()
			logger.Infof("zapd: Checking for updates [%d]", count)
			count += 1
			c <- count
			time.Sleep(time.Hour)
		}

	}()
	upgrade(c, s, updater)
}

func waitUntilOnline() {
	// wait for internet connection
	isOnline := helpers.CheckIfOnline()
	heartbeat := 1
	for isOnline == false {
		time.Sleep(time.Second * time.Duration(heartbeat))
		if heartbeat < 300 {
			heartbeat = heartbeat * 2
		}
		logger.Infof("Not connected to internet, retrying in %d seconds", heartbeat)
		isOnline = helpers.CheckIfOnline()
	}
}
