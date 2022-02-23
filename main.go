package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

var url string = ""

const text = "Inspiration Quote Here!"

func main() {
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/image", parseImageURL)
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	fmt.Println("Server listening on port 3000")
	log.Panic(
		http.ListenAndServe(":3000", nil),
	)

}

type url_struct struct {
	Url string
}

func parseImageURL(w http.ResponseWriter, r *http.Request) {
	//decode response
	decoder := json.NewDecoder(r.Body)

	var URL url_struct
	err := decoder.Decode(&URL)
	checkError(err)

	url = URL.Url
	fmt.Println("URL received!")
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {

	if url != "" {
		//get the http response from the url
		res, err := http.Get(url)
		checkError(err)

		//grab the image from the response body
		data, err := ioutil.ReadAll(res.Body)
		checkError(err)

		img, _, err := image.Decode(bytes.NewReader(data))
		checkError(err)

		res.Body.Close()

		// get image size
		imgWidth := img.Bounds().Dx()
		imgHeight := img.Bounds().Dy()

		//load in a default font
		font, err := truetype.Parse(goregular.TTF)
		checkError(err)

		face := truetype.NewFace(font, &truetype.Options{Size: 48})

		// create canvas for image & drawing text
		dc := gg.NewContext(imgWidth, imgHeight)
		dc.DrawImage(img, 0, 0)
		dc.SetFontFace(face)
		dc.SetColor(color.White)

		// x/y position of text
		x := float64(imgWidth / 2)
		y := float64(imgHeight / 2)
		maxWidth := float64(imgWidth - 60) //maximum width text can occupy

		dc.DrawStringWrapped(text, x, y, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)
		dc.SavePNG("images/inspirational_image.png")

		fmt.Fprintf(w, "<title>Inspirationifier</title>")
		fmt.Fprintf(w, "<h1>Inspirationifier</h1>")
		fmt.Fprintf(w, "<img src='images/inspirational_image.png' style='width:480px;height:480px;'>")
	} else {
		fmt.Fprintf(w, "<title>Inspirationifier</title>")
		fmt.Fprintf(w, "<h1>Inspirationifier</h1>")
		fmt.Fprintf(w, "<p>Please POST an image URL</p>")
	}
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}
