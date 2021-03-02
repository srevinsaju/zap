package daemon

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/srevinsaju/zap/appimage"
	"github.com/srevinsaju/zap/config"
	"github.com/srevinsaju/zap/logging"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var logger = logging.GetLogger()

func upgrade(c chan int, s chan os.Signal, config config.Store) {

	for {
		select {
		case <- c:
			apps, _ := appimage.Upgrade(config)
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

func Sync(config config.Store) {

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
	upgrade(c, s, config)
}

func waitUntilOnline() {
	// wait for internet connection
	isOnline := checkIfOnline()
	heartbeat := 1
	for isOnline == false {
		time.Sleep(time.Second * time.Duration(heartbeat))
		if heartbeat < 300 {
			heartbeat = heartbeat * 2
		}
		logger.Infof("Not connected to internet, retrying in %d seconds", heartbeat)
		isOnline = checkIfOnline()
	}
}

func checkIfOnline() bool {
	// https://dev.to/obnoxiousnerd/check-if-user-is-connected-to-the-internet-in-go-1hk6

	//Make a request to icanhazip.com
	//We need the error only, nothing else :)
	_, err := http.Get("https://icanhazip.com/")
	//err = nil means online
	if err == nil {
		return true
	}
	//if the "return statement" in the if didn't executed,
	//this one will execute surely
	return false
}
