package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"github.com/unrolled/render" // or "gopkg.in/unrolled/render.v1"
)

type ImageRender struct {
	BinaryPath *string
}

func (r *ImageRender) BuildImageOptions(req *http.Request, format string) (ImageOptions, error) {
	url := req.Form.Get("url")

	var html string
	if len(url) == 0 {
		html = req.Form.Get("html")

		if len(html) == 0 {
			return ImageOptions{}, errors.New("url can't be null")
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
	return c, nil
}

//render image bytes to browser
func (r *ImageRender) RenderBytes(w http.ResponseWriter, req *http.Request, format string) {
	err := req.ParseForm()
	if err != nil {
		log.Println("parse form err: ", err)
		return
	}
	c, err := r.BuildImageOptions(req, format)
	if err != nil {
		w.Write([]byte(fmt.Sprint(err)))
		return
	}
	out, err := GenerateImage(&c)
	if err != nil {
		w.Write([]byte(fmt.Sprint(err)))
		return
	}
	w.Write(out)
}

func (r *ImageRender) RenderJson(httpRender *render.Render, w http.ResponseWriter,
	req *http.Request, imgRootDir *string) {
	err := req.ParseForm()
	if err != nil {
		log.Println("parse form err: ", err)
		httpRender.Text(w, http.StatusInternalServerError, fmt.Sprint(err))
		return
	}
	format := req.Form.Get("format")
	if len(format) == 0 {
		httpRender.JSON(w, http.StatusOK,
			map[string]interface{}{"code": 400, "message": "format can't be null"})
		return
	}
	if format != "png" && format != "jpg" {
		httpRender.JSON(w, http.StatusOK,
			map[string]interface{}{"code": 400, "message": "format type invalid"})
		return
	}
	c, err := r.BuildImageOptions(req, format)
	if err != nil {
		httpRender.Text(w, http.StatusInternalServerError, fmt.Sprint(err))
		return
	}
	imgPath := contentToMd5(c.Input+c.Html) + "." + format
	c.Output = *imgRootDir + imgPath
	log.Println("generate file path:", c.Output)
	if !checkFileIsExist(c.Output) {
		_, err = GenerateImage(&c)
		if err != nil {
			httpRender.Text(w, http.StatusInternalServerError, fmt.Sprint(err))
			return
		}
	}

	httpRender.JSON(w, http.StatusOK,
		map[string]interface{}{"code": 200, "url": imgPath})
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func contentToMd5(content string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(content))
	cipherBytes := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherBytes)
}

func main() {
	binPath := flag.String("path", "/usr/local/bin/wkhtmltoimage", "wkhtmltoimage bin path")
	imgRootDir := flag.String("img.dir", "./tmp/", "generated image local dir")
	port := flag.String("web.port", "8080", "web server port")
	flag.Parse()
	imgRender := ImageRender{}
	imgRender.BinaryPath = binPath
	r := render.New()
	mux := http.NewServeMux()
	os.MkdirAll(*imgRootDir, 0755)
	staticHandler := http.FileServer(http.Dir(*imgRootDir))

	mux.HandleFunc("/to/img.png", func(w http.ResponseWriter, req *http.Request) {
		imgRender.RenderBytes(w, req, "png")
	})
	mux.HandleFunc("/to/img.jpg", func(w http.ResponseWriter, req *http.Request) {
		imgRender.RenderBytes(w, req, "jpg")
	})
	mux.HandleFunc("/show/img/", func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = req.URL.Path[9:]
		staticHandler.ServeHTTP(w, req)
	})
	mux.HandleFunc("/api/v1/to/img.json", func(w http.ResponseWriter, req *http.Request) {
		imgRender.RenderJson(r, w, req, imgRootDir)
	})
	if len(*port) == 0 {
		*port = "8080"
	}
	http.ListenAndServe(":"+*port, mux)
}
