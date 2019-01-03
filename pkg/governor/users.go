package governor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/mux"
)

type AuthorizedKeyRequest struct {
	Key string `json:"key"`
}

type AuthorizedKeysResponse struct {
	Keys []string `json:"keys"`
}

var SshKeyRegexp = regexp.MustCompile(`^ssh-\w+ \S+( \S+)?$`)

func NewUsersResource(prefix string) http.Handler {
	router := mux.NewRouter()
	sub := router.PathPrefix(prefix).Subrouter()

	sub.HandleFunc("/{user}/authorized_keys", handleListAuthorizedKeys).Methods("GET")
	sub.HandleFunc("/{user}/authorized_keys", handleAddAuthorizedKey).Methods("POST")

	return router
}

func ListAuthorizedKeys(user string) ([]string, error) {
	var keys []string

	path := fmt.Sprintf("/home/%s/.ssh/authorized_keys", user)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keys = append(keys, scanner.Text())
	}

	return keys, nil
}

func SaveAuthorizedKeys(keys []string, user string) error {
	path := fmt.Sprintf("/home/%s/.ssh/authorized_keys", user)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, key := range keys {
		fmt.Fprintf(file, "%s\n", key)
	}

	return nil
}

func handleAddAuthorizedKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	decoder := json.NewDecoder(r.Body)
	var request AuthorizedKeyRequest
	err := decoder.Decode(&request)
	if err != nil || !SshKeyRegexp.MatchString(request.Key) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	keys, _ := ListAuthorizedKeys(vars["user"])

	for _, key := range keys {
		if key == request.Key {
			response := AuthorizedKeysResponse{
				Keys: keys,
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	keys = append(keys, request.Key)
	err = SaveAuthorizedKeys(keys, vars["user"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := AuthorizedKeysResponse{
		Keys: keys,
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleListAuthorizedKeys(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	keys, err := ListAuthorizedKeys(vars["user"])
	if err == nil {
		response := AuthorizedKeysResponse{
			Keys: keys,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
