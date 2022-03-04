package main

import (
	"TextOverImg/internal"
	"TextOverImg/store"
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
	"os"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/lib/pq"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/gorilla/mux"

	"github.com/gofrs/uuid"
)

type request struct {
	Url  string `json:"url"`
	Text string `json:"text"`
	Auth bool   `json:"auth"`
}

type user struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

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

	//Route handling and endpoints
	r.HandleFunc("/image", createImage).Methods(http.MethodPost)
	r.Handle("/login", login(db)).Methods(http.MethodPost)
	r.Handle("/register", register(db)).Methods(http.MethodPost)

	//Image file handler
	fs := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/image/").Handler(http.StripPrefix("/image/", fs))

	//Static file handler (frontend)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("frontend/dist/")))
	r.PathPrefix("/").HandlerFunc(index("frontend/dist/index.html"))

	//Server
	fmt.Println("Server listening on port 3000")

	err = http.ListenAndServe(":3000", r)
	if err != nil && err != http.ErrServerClosed {
		log.Println(err)
		os.Exit(1)
	}
}

func index(entrypoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, entrypoint)
	}
}

type registerResp struct {
	Msg string
}

func register(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//decode response
		decoder := json.NewDecoder(r.Body)

		var req user
		err := decoder.Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := &registerResp{
			Msg: registerUserInDb(db, req.Username, req.Password),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

type loginResp struct {
	Status string
	Msg    string
	User   string
}

func login(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//decode response
		decoder := json.NewDecoder(r.Body)

		var req user
		err := decoder.Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		querier := store.New(db)
		user, err := querier.GetUser(r.Context(), req.Username)
		if errors.Is(err, sql.ErrNoRows) || !internal.CheckPasswordHash(req.Password, user.PasswordHash) {
			resp := &loginResp{
				Status: "fail",
				Msg:    "Login failed, wrong username or password",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error looking up user", err)
			return
		}

		//valid user
		resp := &loginResp{
			Status: "success",
			Msg:    "Login successful!",
			User:   req.Username,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

type imageResp struct {
	Image string
	Error string
}

func createImage(w http.ResponseWriter, r *http.Request) {
	//decode response
	decoder := json.NewDecoder(r.Body)

	var req request
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//check the request
	if req.Url == "" || req.Text == "" {
		resp := &imageResp{
			Error: "Error - Incomplete Request",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("URL and text received, URL=%v, Text=%v", req.Url, req.Text)

	//get http response from url
	res, err := http.Get(req.Url)
	if err != nil {
		resp := &imageResp{
			Error: "Error - Invalid URL",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	//grab the image from the response body
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Body.Close()

	//generate UUID
	uuid, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//place text over img
	err = textOverImg(data, req.Text, req.Auth, uuid.String())
	if err != nil {
		resp := &imageResp{
			Error: "Error - Could not get image from URL",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	//send the image to the user
	resp := &imageResp{
		Image: "http://localhost:3000/image/inspirational_image_" + uuid.String() + ".png",
		Error: "none",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

//Helper Functions

func textOverImg(imgData []byte, text string, premium bool, uuid string) error {
	//decode from []byte to image.Image
	img, _, err := image.Decode(bytes.NewReader(imgData))
	//if the url is not an image
	if err != nil {
		//fmt.Println("Could not get image from URL.")
		return err
	}
	//get image size
	imgWidth := img.Bounds().Dx()
	imgHeight := img.Bounds().Dy()

	//load in a default font
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Println(err)
		return err
	}

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
	file := "images/inspirational_image_" + uuid + ".png"
	err = dc.SavePNG(file)
	if err != nil {
		return err
	}
	log.Println("Inspirational image created: .", file)
	return nil
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
			return "Error - User already exists"
		}
		if err != nil {
			log.Println("Failed to create user: ", err)
			return "Error - Failed to create user"
		}
		return "Account created! Proceed to the home page and login."
	} else {
		if username == "" && password == "" {
			return "Error - Please enter a username and password"
		} else if username == "" {
			return "Error - Please enter a username"
		} else {
			return "Error - Please enter a password"
		}
	}
}
