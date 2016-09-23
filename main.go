package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	//	"github.com/unrolled/render" // or "gopkg.in/unrolled/render.v1"
)

type ImageRender struct {
	BinaryPath *string
}

//render image with wkhtmltoimage with url
func (r *ImageRender) GetBytes(req *http.Request, format string) ([]byte, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, err
	}
	url := req.Form.Get("url")

	var html string
	if len(url) == 0 {
		html = req.Form.Get("html")

		if len(html) == 0 {
			return nil, errors.New("url can't be null")
		} else {
			url = "-"
			log.Println("render for: ", html)
		}
	} else {
		html = ""
		log.Println("render for: ", url)
	}

	c := ImageOptions{BinaryPath: *r.BinaryPath,
		Input: url, Html: html, Format: format}

	width, err := strconv.Atoi(req.Form.Get("width"))
	if err == nil {
		c.Width = width
	}

	height, err := strconv.Atoi(req.Form.Get("height"))
	if err == nil {
		c.Height = height
	}

	quality, err := strconv.Atoi(req.Form.Get("quality"))
	if err == nil {
		c.Quality = quality
	}

	return GenerateImage(&c)
}

//render image bytes to browser
func (r *ImageRender) RenderBytes(w http.ResponseWriter, req *http.Request, format string) {
	out, err := r.GetBytes(req, format)
	if err != nil {
		w.Write([]byte(fmt.Sprint(err)))
		return
	}
	w.Write(out)
}

func main() {
	binPath := flag.String("path", "/usr/local/bin/wkhtmltoimage", "wkhtmltoimage bin path")
	port := flag.String("web.port", "8080", "web server port")
	flag.Parse()
	render := ImageRender{}
	render.BinaryPath = binPath
	//	r := render.New()
	mux := http.NewServeMux()

	mux.HandleFunc("/render.png", func(w http.ResponseWriter, req *http.Request) {
		render.RenderBytes(w, req, "png")
	})
	mux.HandleFunc("/render.jpg", func(w http.ResponseWriter, req *http.Request) {
		render.RenderBytes(w, req, "jpg")
	})
	if len(*port) == 0 {
		*port = "8080"
	}
	http.ListenAndServe(":"+*port, mux)
}
