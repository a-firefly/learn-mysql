package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		http.Error(w, "email already exists", http.StatusConflict)
		return
	}

	log.Printf("database error: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return 0, false
	}

	return  id, true
}

func readInput(w http.ResponseWriter, r *http.Request) (UserInput, bool) {
	var in UserInput

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&in); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return in, false
	}

	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)

	if in.Name == "" || in.Email == "" {
		http.Error(w, "name and email are required", http.StatusBadRequest)
		return in, false
	}

	return in, true
}