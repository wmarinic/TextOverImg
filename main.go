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
	"strconv"

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

var i int = 0

func main() {
	//Init router
	r := mux.NewRouter()

	// Route handling and endpoints
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("frontend/dist/")))
	r.PathPrefix("/").HandlerFunc(IndexHandler("frontend/dist/index.html"))

	r.HandleFunc("/image", createInspImage).Methods("POST")
	fs := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/image/").Handler(http.StripPrefix("/image/", fs))

	r.HandleFunc("/user", userLogin).Methods("POST")
	r.HandleFunc("/logout", userLogout).Methods("GET")

	fmt.Println("Server listening on port 3000")
	log.Panic(
		http.ListenAndServe(":3000", r),
	)
}

func IndexHandler(entrypoint string) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, entrypoint)
	}
	return http.HandlerFunc(fn)
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
		if err != nil {
			//fmt.Println("Invalid URL.")
			fmt.Fprintf(w, `{"error":"Error: Invalid URL"}`)
		} else {
			//grab the image from the response body
			data, err := ioutil.ReadAll(res.Body)
			checkError(err)

			res.Body.Close()

			//place text over img
			if textOverImg(data, premium) {
				fmt.Fprintf(w, `{"image": "http://localhost:3000/image/inspirational_image_%s.png", "error":"none"}`, strconv.Itoa(i))
			} else {
				//no image from the given url
				fmt.Fprintf(w, `{"error":"Error: Could not get image from URL"}`)
			}
		}
	} else {
		//fmt.Println("Error: Incomplete request.")
		fmt.Fprintf(w, `{"error":"Error: Incomplete request"}`)
	}
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
		//fmt.Println("Login successful!")
		fmt.Fprintf(w, `{"status": "success", "user":"%s", "msg":"Login successful!"}`, userName)
	} else {
		//fmt.Println("Login failed, wrong username or password.")
		fmt.Fprintf(w, `{"status": "fail", "msg":"Login failed, wrong username or password"}`)
	}
}

func userLogout(w http.ResponseWriter, r *http.Request) {
	premium = false
	fmt.Println("User logged out.")
}

//Helper Functions

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func textOverImg(imgData []byte, premium bool) bool {
	//increment image count
	i++
	//decode from []byte to image.Image
	img, _, err := image.Decode(bytes.NewReader(imgData))
	//if the url is not an image
	if err != nil {
		//fmt.Println("Could not get image from URL.")
		return false
	} else {
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
		dc.SavePNG("images/inspirational_image_" + strconv.Itoa(i) + ".png")
		fmt.Println("Inspirational image created.")
		return true
	}
}
