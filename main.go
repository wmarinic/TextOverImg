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

	"github.com/gorilla/mux"
)

//define global vars
var url string = ""
var text string = ""
var premium bool = false

type request_struct struct {
	Url  string `json:"url"`
	Text string `json:"text"`
}

type user_struct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	//Init router
	r := mux.NewRouter()

	// Route handling and endpoints
	r.HandleFunc("/", homePageHandler).Methods("GET")
	r.HandleFunc("/image", dispImage).Methods("GET")
	r.HandleFunc("/image", createInspImage).Methods("POST")
	r.HandleFunc("/user", userLogin).Methods("POST")
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	fmt.Println("Server listening on port 3000")
	log.Panic(
		http.ListenAndServe(":3000", r),
	)
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<title>Inspirationifier</title>")
	fmt.Fprintf(w, "<h1>Inspirationifier</h1>")
	fmt.Fprintf(w, "<h2>Welcome to the Inspirationfier app!</h2>")
	fmt.Fprintf(w, "<p>Please send a POST request with an image URL and the desired text.</p>")
}

func userLogin(w http.ResponseWriter, r *http.Request) {
	//decode response
	decoder := json.NewDecoder(r.Body)

	var user user_struct
	err := decoder.Decode(&user)
	checkError(err)

	userName := user.Username
	passWord := user.Password

	//hard coding a log in for now @TODO: add db + secure pw storing
	if userName == "test" && passWord == "test" {
		//premium access granted
		premium = true
		fmt.Println("Login successful!")
	} else {
		fmt.Println("Login failed, wrong username or password.")
	}
}

func createInspImage(w http.ResponseWriter, r *http.Request) {
	//decode response
	decoder := json.NewDecoder(r.Body)

	var req request_struct
	err := decoder.Decode(&req)
	checkError(err)

	url = req.Url
	text = req.Text

	//check the request
	if url != "" && text != "" {
		fmt.Println("URL and text received.")

		//get http response from url
		res, err := http.Get(url)
		checkError(err)

		//grab the image from the response body
		data, err := ioutil.ReadAll(res.Body)
		checkError(err)

		res.Body.Close()

		//place text over img
		textOverImg(data, premium)
		fmt.Println("Inspirational image created.")
	} else {
		fmt.Println("Error: Incomplete request.")
	}
}

func dispImage(w http.ResponseWriter, r *http.Request) {
	//check if url exists
	if url != "" {
		//check if text exists
		if text != "" {
			fmt.Fprintf(w, "<title>Inspirationifier</title>")
			fmt.Fprintf(w, "<h1>Inspirationifier</h1>")
			fmt.Fprintf(w, "<img src='images/inspirational_image.png' style='width:480px;height:480px;'>")
		} else {
			fmt.Fprintf(w, "<title>Inspirationifier</title>")
			fmt.Fprintf(w, "<h1>Inspirationifier</h1>")
			fmt.Fprintf(w, "<p> Error: Incomplete request, no text found.")
		}
	} else {
		fmt.Fprintf(w, "<title>Inspirationifier</title>")
		fmt.Fprintf(w, "<h1>Inspirationifier</h1>")
		fmt.Fprintf(w, "<p>Please POST an image URL and text</p>")
	}
}

//Helper Functions

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func textOverImg(imgData []byte, premium bool) {
	//decode from []byte to image.Image
	img, _, err := image.Decode(bytes.NewReader(imgData))
	checkError(err)

	//get image size
	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	//load in a default font
	font, err := truetype.Parse(goregular.TTF)
	checkError(err)

	face := truetype.NewFace(font, &truetype.Options{Size: 48})

	//create canvas for image & drawing text
	dc := gg.NewContext(imgWidth, imgHeight)
	dc.DrawImage(img, 0, 0)
	dc.SetFontFace(face)
	dc.SetColor(color.White)

	//set x/y position of text
	x := float64(imgWidth / 2)
	y := float64(imgHeight / 2)
	maxWidth := float64(imgWidth - 60) //maximum width text can occupy

	dc.DrawStringWrapped(text, x, y, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)
	//check users access
	if !premium {
		//draw a watermark
		dc.DrawStringAnchored("Inspirationifier: Free Version.", 325, y*2-48, 0.5, 0.5)
	}
	dc.SavePNG("images/inspirational_image.png")

}
