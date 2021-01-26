package main

type ZapConfig struct {
	mirror string
}

var ZapDefaultConfig = ZapConfig{mirror: "https://g.srevinsaju.me/get-appimage/%s/core.json"}
