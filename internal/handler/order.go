package handler

import (
	"fmt"
	"net/http"
)

type Order struct {
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create an order")
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List all ordres")
}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get an ordre by ID")
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an ordre by ID")
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an ordre by ID")
}
