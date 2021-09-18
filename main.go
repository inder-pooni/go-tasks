package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
)
// data structure for representing user resource
type user struct {
	ID string `json:"id,omitempty"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Age int `json:"age"`
}

var users = map[string]*user{}

func getUsers(ctx echo.Context) error {
	userSlice := []user{}

	for _ , v := range users {
		userSlice = append(userSlice, *v)
	}
	return ctx.JSON(http.StatusOK , userSlice)
}

func getUserById(ctx echo.Context) error {
	id := ctx.Param("id")
	user := users[id]

	if user == nil {
		errMsg := fmt.Sprintf("Cound not find user with id: %s", id)
		return ctx.JSON(http.StatusNotFound, errMsg)
	}

	return ctx.JSON(http.StatusOK , user)
}

func createUser(ctx echo.Context) error {
	user := user{}
	defer ctx.Request().Body.Close()
	err := json.NewDecoder(ctx.Request().Body).Decode(&user)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}
	user.ID = uuid.New().String()
	log.Print(&user)

	users[user.ID] = &user

	log.Print("Successfully added new User")
	return ctx.JSON(http.StatusOK, &user)
}
// echo context represents the context of the currect http request
func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/users", getUsers)
	e.POST("/users", createUser)
	e.GET("/user/:id", getUserById)
	e.Logger.Fatal(e.Start(":8080"))
}
// endpoints: get users , post user , get user {id}