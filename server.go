package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/nfnt/resize"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strconv"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "image/jpeg")
	fmt.Println(r.URL.Path)
	fmt.Println("into imagehandler")
	http.ServeFile(w, r, r.URL.Path[1:])
}

func main() {
	m := martini.Classic()
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
				//return 404, "Not Found"
			}
			// decode jpeg into image.Image
			img, err := jpeg.Decode(file)
			if err != nil {
				log.Fatal(err)
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
				log.Fatal(err)
			}
			log.Printf(out.Name())
			defer out.Close()

			// write new image to file
			jpeg.Encode(out, m, nil)
			http.ServeFile(res, req, out.Name())
		}
	})
	m.RunOnAddr(":8080")

}
