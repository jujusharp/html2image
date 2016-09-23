package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	//	"github.com/unrolled/render" // or "gopkg.in/unrolled/render.v1"
)

//render image with wkhtmltoimage with url
func renderImage(req *http.Request, format string, binPath string) ([]byte, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, err
	}
	url := req.Form.Get("url")

	if len(url) == 0 {
		return nil, errors.New("url can't be null")
	}
	log.Println("render for: ", url)
	c := ImageOptions{BinaryPath: binPath,
		Input: url, Format: format}
	return GenerateImage(&c)
}

func main() {
	binPath := flag.String("path", "/usr/local/bin/wkhtmltoimage", "wkhtmltoimage bin path")
	port := flag.String("web.port", "8080", "web server port")
	flag.Parse()
	//	r := render.New()
	mux := http.NewServeMux()

	mux.HandleFunc("/render.png", func(w http.ResponseWriter, req *http.Request) {
		out, err := renderImage(req, "png", *binPath)
		if err != nil {
			w.Write([]byte(fmt.Sprint(err)))
			return
		}
		w.Write(out)
	})
	mux.HandleFunc("/render.jpg", func(w http.ResponseWriter, req *http.Request) {
		out, err := renderImage(req, "jpg", *binPath)
		if err != nil {
			w.Write([]byte(fmt.Sprint(err)))
			return
		}
		w.Write(out)
	})
	if len(*port) == 0 {
		*port = "8080"
	}
	http.ListenAndServe(":"+*port, mux)
}
