package ua

import "github.com/mssola/useragent"

type Info struct {
	Browser string
	OS      string
	Device  string
}

func Parse(uaString string) Info {
	parsed := useragent.New(uaString)
	browser, _ := parsed.Browser()

	device := "desktop"
	if parsed.Mobile() {
		device = "mobile"
	}

	return Info{
		Browser: browser,
		OS:      parsed.OS(),
		Device:  device,
	}
}
