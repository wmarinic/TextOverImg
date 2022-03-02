package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"TextOverImg/internal"
	"TextOverImg/store"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type request_struct struct {
	Url  string `json:"url"`
	Text string `json:"text"`
	Auth bool   `json:"auth"`
}

type user_struct struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

var i int = 0

func main() {
	//init db
	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		"local",
		"pass",
		"localhost",
		5432,
		"inspirationifierdb",
	)
	//Open db
	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		log.Fatalln("Error opening db: ", err)
	}
	//Check connection
	if err := db.Ping(); err != nil {
		log.Fatalln("Error from db ping: ", err)
	}

	//Create a test user
	createUserInDb(db)

	//Init router
	r := mux.NewRouter()

	// Route handling and endpoints
	r.HandleFunc("/image", createInspImage).Methods("POST")
	fs := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/image/").Handler(http.StripPrefix("/image/", fs))
	r.Handle("/login", userLogin(db)).Methods("POST")
	r.Handle("/register", userRegister(db)).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("frontend/dist/")))
	r.PathPrefix("/").HandlerFunc(IndexHandler("frontend/dist/index.html"))
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

func userRegister(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//decode response
		decoder := json.NewDecoder(r.Body)

		var register_req user_struct
		err := decoder.Decode(&register_req)
		checkError(err)

		userName := register_req.Username
		passWord := register_req.Password

		msg := registerUserInDb(db, userName, passWord)
		fmt.Fprint(w, msg)
	})
}

func userLogin(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//decode response
		decoder := json.NewDecoder(r.Body)

		var login_req user_struct
		err := decoder.Decode(&login_req)
		checkError(err)

		userName := login_req.Username
		passWord := login_req.Password

		querier := store.New(db)

		user, err := querier.GetUser(r.Context(), userName)
		if errors.Is(err, sql.ErrNoRows) || !internal.CheckPasswordHash(passWord, user.PasswordHash) {
			fmt.Fprintf(w, `{"status": "fail", "msg":"Login failed, wrong username or password"}`)
			return
		}
		if err != nil {
			fmt.Println("Error looking up user", err)
			return
		}

		//valid user
		log.Println("Login successful!")
		fmt.Fprintf(w, `{"status": "success", "user":"%s", "msg":"Login successful!"}`, userName)
	})
}

func createInspImage(w http.ResponseWriter, r *http.Request) {
	//decode response
	decoder := json.NewDecoder(r.Body)

	var req request_struct
	err := decoder.Decode(&req)
	checkError(err)

	url := req.Url
	text := req.Text
	auth := req.Auth

	//log.Print("premium access: ")
	//log.Println(auth)

	//check the request
	if url != "" && text != "" {
		log.Println("URL and text received.")

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
			if textOverImg(data, text, auth) {
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

//Helper Functions

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func textOverImg(imgData []byte, text string, premium bool) bool {
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
		log.Println("Inspirational image created.")
		return true
	}
}

func createUserInDb(db *sql.DB) {
	ctx := context.Background()

	querier := store.New(db)

	log.Println("Creating test user...")
	hashPwd := internal.HashPassword("test")

	_, err := querier.CreateUser(ctx, store.CreateUserParams{
		UserName:     "test",
		PasswordHash: hashPwd,
	})

	if err, ok := err.(*pq.Error); ok && err.Code.Name() == "unique_violation" {
		log.Println("Test user already exists")
		return
	}
	if err != nil {
		log.Println("Failed to create user: ", err)
	}
}

func registerUserInDb(db *sql.DB, username, password string) string {
	//check user and pw
	if username != "" && password != "" {
		ctx := context.Background()

		querier := store.New(db)

		log.Println("Creating new user...")
		hashPwd := internal.HashPassword(password)

		_, err := querier.CreateUser(ctx, store.CreateUserParams{
			UserName:     username,
			PasswordHash: hashPwd,
		})

		if err, ok := err.(*pq.Error); ok && err.Code.Name() == "unique_violation" {
			log.Println("User already exists")
			return `{"msg": "Error: User already exists"}`
		}
		if err != nil {
			log.Println("Failed to create user: ", err)
			return `{"msg": "Error: Failed to create user"}`
		}
		return `{"msg": "Account created! Proceed to the home page and login."}`
	} else {
		if username == "" && password == "" {
			return `{"msg": "Error: Please enter a username and password"}`
		} else if username == "" {
			return `{"msg": "Error: Please enter a username"}`
		} else {
			return `{"msg": "Error: Please enter a password"}`
		}

	}

}
