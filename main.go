package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type todo struct {
	Title       string
	Description string
	Complete    bool
}

type update struct {
	Old todo
	New todo
}

type description struct {
	Description string `json:"description"`
}

var cache = make(map[string]todo)

func getAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(cache); err != nil {
		panic(err)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	if title, exists := pathParams["title"]; exists {
		obj := cache[title]
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			panic(err)
		}
	}
	w.Header().Set("Content-Type", "application/json")
}

func post(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	if title, exists := pathParams["title"]; exists {
		var d description
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			panic(err)
		}
		if err := r.Body.Close(); err != nil {
			panic(err)
		}
		if err := json.Unmarshal(body, &d); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(422)
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
		}
		t := createTodo(title, d)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(t); err != nil {
			panic(err)
		}
	}
}

func getBody(w http.ResponseWriter, r *http.Request) (todo, error) {
	var body todo
	raw, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return body, err
	}
	if err := r.Body.Close(); err != nil {
		return body, err
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}
	return body, nil
}

func createTodo(t string, d description) todo {
	td := todo{
		t,
		d.Description,
		false,
	}
	cache[t] = td
	return td
}

func put(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	body, err := getBody(w, r)
	if err != nil {
		panic(err)
	}

	var p update
	if title, exists := pathParams["title"]; exists {
		old, ok := cache[title]
		if ok {
			cache[title] = body
			p = update{old, body}
		} else {
			cache[title] = body
			p = update{old, body}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(p); err != nil {
			panic(err)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"message": "put called"}`))
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	if title, exists := pathParams["title"]; exists {
		if t, ok := cache[title]; ok {
			delete(cache, title)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(t); err != nil {
				panic(err)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "delete called"}`))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`404 Not Found`))
}

func main() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/items", getAll).Methods(http.MethodGet)
	api.HandleFunc("/items/{title}", get).Methods(http.MethodGet)
	api.HandleFunc("/items/{title}", post).Methods(http.MethodPost)
	api.HandleFunc("/items/{title}", put).Methods(http.MethodPut)
	api.HandleFunc("/items/{title}", deleteHandler).Methods(http.MethodDelete)
	api.HandleFunc("", notFound)
	log.Fatal(http.ListenAndServe(":8080", r))
}
