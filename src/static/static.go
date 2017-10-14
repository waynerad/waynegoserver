package static

import (
	"io"
	"net/http"
	"os"
)

func FigureOutContentTypeFromFilename(filename string) string {
	if len(filename) > 3 {

		if filename[len(filename)-4:] == ".pdf" {
			return "application/pdf"
		} else {
			if filename[len(filename)-4:] == ".css" {
				return "text/css; charset=utf-8"
			} else {
				if filename[len(filename)-4:] == ".mp3" {
					return "audio/mpeg"
				} else {
					if filename[len(filename)-4:] == ".ico" {
						return "image/vnd.microsoft.icon"
					} else {
						return "text/html; charset=utf-8"
					}
				}
			}
		}

	}
	return "text/html; charset=utf-8"
}

func OutputStaticFileWithContentType(w http.ResponseWriter, filename string) {
	contentType := FigureOutContentTypeFromFilename(filename)
	fh, err := os.Open(filename)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	header := w.Header()
	header.Set("Content-Type", contentType)
	_, err = io.Copy(w, fh)
	fh.Close()
}

func StaticHandler(w http.ResponseWriter, r *http.Request, operation string) {
	if operation == "/" {
		operation = "/index.html"
	}
	filename := "/home/ec2-user/wayneserver/staticdoc/" + operation[1:]
	// fmt.Println("filename", filename)
	OutputStaticFileWithContentType(w, filename)
}

func StaticDirHandler(w http.ResponseWriter, r *http.Request, dir string, operation string) {
	// fmt.Println("dir", dir)
	if operation == "/" {
		operation = "/index.html"
	}
	filename := "/home/ec2-user/wayneserver/staticdoc/" + dir + "/" + operation[1:]
	// fmt.Println("filename", filename)
	OutputStaticFileWithContentType(w, filename)
}
