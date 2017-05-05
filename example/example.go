package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kere/phantomjs"
)

func main() {
	p := phantomjs.NewProcess()
	p.BinPath = "phantomjs.exe"
	p.AddOption(`--config="D:/Server/phantomjs/conf/spider.json"`)

	// Start the process once.
	if err := p.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer p.Close()

	// Do other stuff in your program.
	page, err := p.CreateWebPage()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer page.Close()

	settings, err := page.Settings()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	settings.JavascriptEnabled = true
	settings.LoadImages = false
	settings.LocalToRemoteURLAccessEnabled = false
	settings.WebSecurityEnabled = false
	settings.ResourceTimeout = time.Second * 6
	settings.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36"

	page.SetSettings(settings)

	// Open a URL.
	// url := "https://kere.github.io"
	url := "http://localhost:8080/calc"
	if err = page.Open(url); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if err = page.InjectJS("jquery.min.js"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	info, err := page.Evaluate(`function() {
    // return $('#txtMethod').val();
  }`)
	if err != nil {
		fmt.Println("eva:", err)
		os.Exit(1)
	}

	page.SetViewportSize(1000, 800)
	if err = page.Render("page.png", "png", 100); err != nil {
		fmt.Println("eva:", err)
		os.Exit(1)
	}

	fmt.Println(info)
}
