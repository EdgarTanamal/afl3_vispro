package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

type Recipe struct {
	ID           int    `json:"ID"`
	Name         string `json:"Name"`
	Ingredients  string `json:"Ingredients"`
	Instructions string `json:"Instructions"`
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root@tcp(localhost:3306)/recipe")
	if err != nil {
		log.Fatal(err)
	}
}

func getAllRecipes(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query("SELECT ID, Name, Ingredients, Instructions FROM recipes")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		err := rows.Scan(&recipe.ID, &recipe.Name, &recipe.Ingredients, &recipe.Instructions)
		if err != nil {
			log.Fatal(err)
		}
		recipes = append(recipes, recipe)
	}

	json.NewEncoder(w).Encode(recipes)
}

func getRecipeByID(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	row := db.QueryRow("SELECT ID, Name, Ingredients, Instructions FROM recipes WHERE ID = ?", id)

	var recipe Recipe
	err := row.Scan(&recipe.ID, &recipe.Name, &recipe.Ingredients, &recipe.Instructions)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Resep dengan ID %d tidak ditemukan", id)
			return
		}
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(recipe)
}

func addNewRecipe(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	w.Header().Set("Content-Type", "application/json")
	var newRecipe Recipe
	_ = json.NewDecoder(r.Body).Decode(&newRecipe)

	stmt, err := db.Prepare("INSERT INTO recipes(Name, Ingredients, Instructions) VALUES(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(newRecipe.Name, newRecipe.Ingredients, newRecipe.Instructions)
	if err != nil {
		log.Fatal(err)
	}

	newID, _ := res.LastInsertId()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Resep berhasil ditambahkan", "id": newID})
}

func updateRecipeByID(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	var updatedRecipe Recipe
	_ = json.NewDecoder(r.Body).Decode(&updatedRecipe)

	stmt, err := db.Prepare("UPDATE recipes SET Name=?, Ingredients=?, Instructions=? WHERE ID=?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(updatedRecipe.Name, updatedRecipe.Ingredients, updatedRecipe.Instructions, id)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Resep dengan ID %d tidak ditemukan", id)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Resep berhasil diperbarui"})
}

func deleteRecipeByID(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	stmt, err := db.Prepare("DELETE FROM recipes WHERE ID=?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Resep dengan ID %d tidak ditemukan", id)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Resep berhasil dihapus"})
}

func main() {
	initDB()
	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/recipes", getAllRecipes).Methods("GET")
	router.HandleFunc("/recipes/{id}", getRecipeByID).Methods("GET")
	router.HandleFunc("/recipes/add", addNewRecipe).Methods("POST")
	router.HandleFunc("/recipes/{id}/update", updateRecipeByID).Methods("PUT", "PATCH")
	router.HandleFunc("/recipes/{id}/delete", deleteRecipeByID).Methods("DELETE")

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
