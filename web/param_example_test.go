package web_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/alioygur/gores"
	"github.com/b1naryth1ef/sheath/web"
	"github.com/go-chi/chi/v5"
)

type User struct {
	Id       string
	Username string
}

var userMiddleware, getRequestUser = web.MakeURLParamMiddleware("user", func(value string) (*User, error) {
	return &User{Id: value, Username: "Test"}, nil
})

func ExampleMakeURLParamMiddleware() {
	rtr := chi.NewRouter()
	rtr.With(userMiddleware).Get("/user/{user}", func(w http.ResponseWriter, r *http.Request) {
		gores.JSON(w, http.StatusOK, getRequestUser(r))
	})

	r := httptest.NewRequest("GET", "http://example.com/user/test", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, r)

	resp := w.Result()
	var result User
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s / %s", result.Id, result.Username)
	// Output: test / Test
}
