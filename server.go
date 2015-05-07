package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/gzip"
	"github.com/nfnt/resize"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	m := martini.Classic()
	m.Use(gzip.All())
	m.Get("/", func() string {
		return "Hello world!"
	})

	m.Get("/images/:id/:width", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		imgName := params["id"]
		imageToOpen := "images/" + imgName
		width := params["width"]
		newImgName := width + "__" + imgName

		// equivalent to Python's `if os.path.exists(filename)`
		if _, err := os.Stat("images/" + newImgName); err == nil {
			fmt.Printf("file exists; serving...")
			http.ServeFile(res, req, "images/"+newImgName)
		} else {

			file, err := os.Open(imageToOpen)
			if err != nil {
				log.Printf("Error on opening image file %s", imgName)
				http.Error(res, "Error opening image file", 401)
				//return 404, "Not Found"
				return
			}
			// decode jpeg into image.Image
			img, err := jpeg.Decode(file)
			if err != nil {
				log.Printf("Error on file decode %s", err)
				http.Error(res, "Error on opening image file", 401)
				return
			}
			file.Close()
			log.Printf("Width at %s", width)
			i, err := strconv.ParseUint(width, 0, 64)
			if err != nil {
				log.Printf("error parsing size %s", params["width"])
			}
			m := resize.Resize(uint(i), 0, img, resize.Bilinear)
			out, err := os.Create("images/" + newImgName)
			if err != nil {
				log.Printf("Unable to create new image file %s", err)
			}
			log.Printf(out.Name())
			defer out.Close()

			// write new image to file
			jpeg.Encode(out, m, nil)
			http.ServeFile(res, req, out.Name())
		}
	})
	m.Run()

}
