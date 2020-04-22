# html2image
a golang wrapper to reander html, markdown to image

## Usage

### directly render image

1. render png

[http://127.0.0.1:10000/v1/html2img/to/img.png?url=http://www.google.com](http://127.0.0.1:10000/v1/html2img/to/img.png?url=http://www.google.com)

2. render jpg

[http://127.0.0.1:10000/v1/html2img/to/img.jpg?url=http://www.google.com](http://127.0.0.1:10000/v1/html2img/to/img.jpg?url=http://www.google.com)

3. render png (markdown only)

[http://127.0.0.1:10000/v1/md2img/to/img.png?url=https://github.com/Ink-33/html2image/raw/master/README.md](http://127.0.0.1:10000/v1/md2img/to/img.png?url=https://github.com/Ink-33/html2image/raw/master/README.md)

4. render jpg (markdown only)

[http://127.0.0.1:10000/v1/md2img/to/img.jpg?url=https://github.com/Ink-33/html2image/raw/master/README.md](http://127.0.0.1:10000/v1/md2img/to/img.jpg?url=https://github.com/Ink-33/html2image/raw/master/README.md)
**Notice**: 
This program will analyze _markdown_ (when `url` end with `.md`). If you don't want to render a _markdown_ file, please add the `nomd=true` parameter to the url.  

### render image and return image url info

1. render image and return json

[http://127.0.0.1:10000/v1/html2img/to/img.json?url=http://www.google.com&format=jpg](http://127.0.0.1:10000/v1/html2img/to/img.json?url=http://www.google.com&format=jpg)

2. show image url from the json result

[http://127.0.0.1:10000/v1/html2img/show/img/](http://127.0.0.1:10000/v1/html2img/show/img/){your image url}

## More Params In Url
```shell
 html: the html content to render, if url has set, this param will ignore
 md: the markdown content to render, if url has set,
 this param will ignore (not support yet)
 nomd: ingore markdown url 
 width: the html page width
 height: the html page height
 quality: the image quality
 
```

## Changelog
2020/04/23 Add markdown render support.