package tui

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/srevinsaju/zap/logging"
)

var logger = logging.GetLogger()

// DownloadFileWithProgressBar downloads a file from the internet, with the URL url
// and saves the file in the destination file path 'destination', while showing a
// progress bar in the command line output. name is used to visually show to a user
// what kind of file is being downloaded.
func DownloadFileWithProgressBar(url string, destination string, name string) error {
	logger.Debug("Attempting to do http request")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("The file asset cannot be accessed, possibly it was removed.")
	}

	f, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY, 0755)
  if err != nil {
    panic(err)
  }

	fmt.Printf("Downloading %s\n", name)
	logger.Debug("Setting up progressbar")
	bar := NewProgressBar(
		int(resp.ContentLength),
		"i",
	)
  mw := io.MultiWriter(f, bar)
	_, err = io.Copy(mw, resp.Body)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}
	// need a newline here
	fmt.Print("\n")
	return nil
}
