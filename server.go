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

	m.Get("/size/:width/**", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		reqImageName := params["_1"]
		width := params["width"]
		var (
			resizedImage image.Image
			img          image.Image
		)

		iWidth, err := strconv.ParseUint(width, 0, 64)
		if err != nil {
			log.Printf("error parsing size %s", params["width"])
		}

		fileRequested, err := getFileWithPath(reqImageName)
		log.Printf("Requested file %s", fileRequested)
		if err != nil {
			http.Error(res, "unable to load file", 401)
		}
		// is it local?
		switch checkLocal(fileRequested, width) {
		case BASE_FILE_PRESENT:
			// resize the file and serve
			img, err = loadLocalImage(reqImageName)
			if err != nil {
				http.Error(res, "Error on opening image file", 401)
				return
			}
			resizedImage = getResizedImage(img, iWidth)

		case NO_LOCAL_FILE:
			// load image remotely
			img, err = loadRemoteImage(reqImageName)
			if err != nil {
				http.Error(res, "Error on opening image file", 401)
				return
			}
			resizedImage = getResizedImage(img, iWidth)

		}

		res.Header().Set("Content-Type", "image/jpeg")
		encodeError := jpeg.Encode(res, resizedImage, &jpeg.Options{100})
		if encodeError != nil {
			res.WriteHeader(500)
		} else {
			res.WriteHeader(200)
		}
	})

	m.Run()

}

// resize the image, that's it
func getResizedImage(img image.Image, width uint64) image.Image {
	m := resize.Resize(uint(width), 0, img, resize.Bilinear)
	return m
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
	file, err := os.Open(LOCAL_PREFIX + imgName)
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
// this would change to eventually use a key to map to a url and deny and http on the path
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

// check to see if file is cached locally
func checkLocal(imgName string, width string) int {
	if _, err := os.Stat(LOCAL_PREFIX + imgName); err == nil {
		fmt.Printf("base file exists; serving...")
		return BASE_FILE_PRESENT
	}
	return NO_LOCAL_FILE

}
