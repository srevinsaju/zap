package helpers

import "net/http"

func CheckIfOnline() bool {
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
