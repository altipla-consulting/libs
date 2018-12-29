package funcs

import (
	"encoding/json"
	"fmt"
	"os"
)

var assets = map[string]string{}

func init() {
	f, err := os.Open("rev-manifest.json")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}

		panic(err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&assets); err != nil {
		panic(err)
	}
}

func Asset(assetsHostname, url string) string {
	m, ok := assets[url]
	if ok {
		url = m
	}

	return fmt.Sprintf("%s%s", assetsHostname, url)
}

func Rev(url string) string {
	m, ok := assets[url]
	if ok {
		return m
	}

	return url
}

func Development() bool {
	return os.Getenv("VERSION") == ""
}
