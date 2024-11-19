package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
)

var BACK_ADDRESS string

type ProductBody struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func sendDeleteRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	return client.Do(req)
}

func handleDeleteResponse(resp *http.Response, w http.ResponseWriter) {
	if resp.StatusCode == http.StatusOK {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	url := fmt.Sprintf("%s/delete/%s", BACK_ADDRESS, id)

	resp, err := sendDeleteRequest(url)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	defer resp.Body.Close()

	handleDeleteResponse(resp, w)
}

func CancelHandler(w http.ResponseWriter, r *http.Request) {
	component := NewProductForm()
	component.Render(r.Context(), w)
}

func NewProductHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	description := r.FormValue("description")
	priceStr := r.FormValue("price")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	product := ProductBody{Id: 1, Name: name, Description: description, Price: price}
	url := fmt.Sprintf("%s/new/%s/d=%s/%s", BACK_ADDRESS, product.Name, product.Description, priceStr)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	var p []ProductBody
	if err := json.Unmarshal(body, &p); err != nil {
		log.Fatal("Error unmarshalling response:", err)
	}

	product = p[0]
	component := Product(product)
	w.WriteHeader(http.StatusOK)
	component.Render(r.Context(), w)
}

func EditHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	description := r.FormValue("description")
	priceStr := r.FormValue("price")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	product := ProductBody{Id: id, Name: name, Description: description, Price: price}
	url := fmt.Sprintf("%s/edit/%s/%s/d=%s/%s", BACK_ADDRESS, idStr, name, description, priceStr)

	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	component := Product(product)
	component.Render(r.Context(), w)
	w.WriteHeader(http.StatusOK)
}

// ChangeFormEditHandler prepares the form to edit a product.
func ChangeFormEditHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Panic("Invalid product ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	name := r.PathValue("name")
	description := r.PathValue("description")
	priceStr := r.PathValue("price")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		log.Print("Invalid price format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	product := ProductBody{id, name, description[2:], price}
	component := EditProductForm(product)
	component.Render(r.Context(), w)
}

func StartHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(BACK_ADDRESS + "/products/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	var products []ProductBody
	if err := json.Unmarshal(body, &products); err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}

	component := Page(products)
	component.Render(r.Context(), w)
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	BACK_ADDRESS = os.Getenv("BACK_ADDRESS")
}

func main() {
	loadEnv()

	PORT := os.Getenv("PORT")
	http.HandleFunc("DELETE /delete/{id}/", DeleteHandler)
	http.HandleFunc("GET /cancel/", CancelHandler)
	http.HandleFunc("POST /edit/{id}/", EditHandler)
	http.HandleFunc("GET /edit-product/{id}/{name}/{description}/{price}/", ChangeFormEditHandler)
	http.HandleFunc("POST /new/", NewProductHandler)

	form := NewProductForm()
	http.Handle("GET /new/", templ.Handler(form))

	http.HandleFunc("GET /", StartHandler)

	fmt.Println("Listening on " + PORT)
	http.ListenAndServe(PORT, nil)
}
