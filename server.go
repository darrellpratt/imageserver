package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/gzip"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

const BASE_FILE_PRESENT int = 10
const RESIZED_FILE_PRESENT int = 20
const NO_LOCAL_FILE int = 100
const LOCAL_PREFIX string = "images/"
const DELIM string = "__"

func main() {
	m := martini.Classic()
	m.Use(gzip.All())
	m.Get("/", func() string {
		return "Hello world!"
	})

	// 1. check for image locally first,
	//		a. resize if present and serve if present
	//		b. if not there,
	//			i.  if nothing, pull from origin and save locally
	//			ii. if base file exists, load
	//		c. resize image
	//		d. serve new image
	//
	m.Get("/images/:id/:width", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		reqImageName := params["id"]
		width := params["width"]

		fileRequested, err := getFileWithPath(reqImageName)
		log.Printf("Requested file %s", fileRequested)
		if err != nil {
			http.Error(res, "unable to load file", 401)
		}
		// is it local?
		switch checkLocal(fileRequested, width) {
		case RESIZED_FILE_PRESENT:
			{

			}
		case BASE_FILE_PRESENT:
			{

			}
		case NO_LOCAL_FILE:
			{

			}

		}

		// sizedImageName := width + DELIM + ""

	})

	m.Get("/blah/images/:id/:width", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		imgName := params["id"]
		fmt.Println(path.Base(imgName))
		// if strings.HasPrefix(imgName, "http") {
		// 	//url = Url.Parse
		// 	imgResponse, err := http.Get(imgName)
		// 	if err != nil || imgResponse.StatusCode != 200 {
		// 		http.Error(res, "Unable to fetch remote image", 401)
		// 	}
		// 	defer imgResponse.Body.Close()
		// 	m, err := jpeg.Decode(imgResponse.Body)
		// 	if err != nil {
		// 		// handle error
		// 	}

		// } else {

		// }

		imageToOpen := LOCAL_PREFIX + imgName
		width := params["width"]
		newImgName := width + "__" + imgName

		// equivalent to Python's `if os.path.exists(filename)`
		if _, err := os.Stat(LOCAL_PREFIX + newImgName); err == nil {
			fmt.Printf("file exists; serving...")
			http.ServeFile(res, req, LOCAL_PREFIX+newImgName)
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
			out, err := os.Create(LOCAL_PREFIX + newImgName)
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

// save image on local FS
func saveImageFile(imageFileName string, img image.Image) error {
	out, err := os.Create(LOCAL_PREFIX + imageFileName)
	if err != nil {
		log.Printf("Unable to create new image file %s", err)
		return err
	}
	log.Printf(out.Name())
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, img, nil)
	return nil
}

func loadRemoteImage(imgName string) (image.Image, error) {

	imgResponse, err := http.Get(imgName)
	if err != nil || imgResponse.StatusCode != 200 {
		return nil, err
	}
	defer imgResponse.Body.Close()
	m, err := jpeg.Decode(imgResponse.Body)
	if err != nil {
		return nil, err
	}
	return m, nil

}

func loadLocalImage(imgName string) (image.Image, error) {
	file, err := os.Open(imgName)
	if err != nil {
		log.Printf("Error on opening image file %s", imgName)

		//return 404, "Not Found"
		return nil, err
	}
	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Printf("Error on file decode %s", err)

		return nil, err
	}
	file.Close()
	return img, nil

}

func isRemote(imgName string) bool {
	if strings.HasPrefix(imgName, "http") {
		return true
	} else {
		return false
	}
}

// check to see if requeted image is local or remote. if remote pull path element only for requesting code
func getFileWithPath(imgName string) (string, error) {
	if strings.HasPrefix(imgName, "http") {
		u, err := url.Parse(imgName)
		if err != nil {
			return "", err
		}
		imageWithPath := u.Path
		return imageWithPath, nil
	} else {
		return imgName, nil
	}
}

// check to see if file is cached locally with size first, then base if not by size
// remove width, cache remote
func checkLocal(imgName string, width string) int {
	newImgName := width + "__" + imgName
	if _, err := os.Stat(LOCAL_PREFIX + newImgName); err == nil {
		fmt.Printf("resized file exists; serving...")
		return RESIZED_FILE_PRESENT
	}
	// can't find sized file what about base
	if _, err := os.Stat(LOCAL_PREFIX + imgName); err == nil {
		fmt.Printf("base file exists; serving...")
		return BASE_FILE_PRESENT
	}
	return NO_LOCAL_FILE

}
