package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/unrolled/render" // or "gopkg.in/unrolled/render.v1"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type ImageRender struct {
	BinaryPath *string
}

func md2html(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return "<p>Render Fail. Reason: Can not get content. </p>"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "<p>Render Fail. Reason: Can not read body. </p>"
	}
	return string(blackfriday.Run(body))
}

func (r *ImageRender) BuildImageOptions(req *http.Request, format string) (ImageOptions, error) {
	path := req.URL.Path
	url := req.Form.Get("url")
	reqIP := req.Header.Get("X-Forwarded-For")
	var html string

	if !strings.HasPrefix(path, "/v1/md2img/") {
		if len(url) == 0 {
			html = req.Form.Get("html")

			if len(html) == 0 {
				return ImageOptions{}, errors.New("url can't be null")
			} else {
				url = "-"
				log.Println("render for:", html, " RemoteIP:", reqIP)
			}
		} else if strings.HasSuffix(url, ".md") {
			if req.Form.Get("nomd") != "true" {
				html = "<meta charset=\"utf-8\">" + md2html(url)
				log.Println("render markdown for:", url, " RemoteIP:", reqIP)
				url = "-"
			} else {
				html = ""
				log.Println("render for:", url, " RemoteIP:", reqIP)
			}
		} else {
			html = ""
			log.Println("render for:", url, " RemoteIP:", reqIP)
		}
	} else {
		if strings.HasSuffix(url, ".md") {
			if req.Form.Get("nomd") != "true" {
				html = "<meta charset=\"utf-8\">" + md2html(url)
				log.Println("render markdown for:", url, " RemoteIP:", reqIP)
				url = "-"
			} else {
				log.Println("Prevent \"nomd\" in force markdown mode. url:", url, " RemoteIP:", reqIP)
				return ImageOptions{}, errors.New("\"nomd\" cannot be true in force markdown mode")

			}
		} else {
			log.Println("Prevent rendering ", url, "in force markdown mode.", " RemoteIP:", reqIP)
			return ImageOptions{}, errors.New("url must be end with \".md\"in force markdown mode")
		}
	}

	c := ImageOptions{BinaryPath: *r.BinaryPath,
		Input: url, HTML: html, Format: format}

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

func (r *ImageRender) RenderJSON(httpRender *render.Render, w http.ResponseWriter,
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

	today := time.Now().Format("06/01/02/")
	os.MkdirAll(*imgRootDir+today, 0755)
	imgPath := today + contentToMd5(c.Input+c.HTML) + "." + format
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
		map[string]interface{}{"code": 200, "result": imgPath})
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
	port := flag.String("web.port", "10000", "web server port")
	flag.Parse()
	imgRender := ImageRender{}
	imgRender.BinaryPath = binPath
	r := render.New()
	mux := http.NewServeMux()
	staticHandler := http.FileServer(http.Dir(*imgRootDir))
	log.Println("Start.")
	mux.HandleFunc("/v1/html2img/to/img.png", func(w http.ResponseWriter, req *http.Request) {
		imgRender.RenderBytes(w, req, "png")
	})
	mux.HandleFunc("/v1/html2img/to/img.jpg", func(w http.ResponseWriter, req *http.Request) {
		imgRender.RenderBytes(w, req, "jpg")
	})
	mux.HandleFunc("/v1/html2img/show/img/", func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = req.URL.Path[9:]
		staticHandler.ServeHTTP(w, req)
	})
	mux.HandleFunc("/v1/html2img/to/img.json", func(w http.ResponseWriter, req *http.Request) {
		imgRender.RenderJSON(r, w, req, imgRootDir)
	})
	mux.HandleFunc("/v1/md2img/to/img.png", func(w http.ResponseWriter, req *http.Request) {
		imgRender.RenderBytes(w, req, "png")
	})
	mux.HandleFunc("/v1/md2img/to/img.jpg", func(w http.ResponseWriter, req *http.Request) {
		imgRender.RenderBytes(w, req, "jpg")
	})
	if len(*port) == 0 {
		*port = "10000"
	}
	http.ListenAndServe(":"+*port, mux)
}
