package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"os"
)

type User struct {
	ID string `json:"id,omitempty"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Age int `json:"age"`
}

type UserSchema struct {
	ID int64
	UUID string
	FirstName string
	LastName string
	Age int64
}

var (
	db *sql.DB
)

func dbConfig() error {
	config := mysql.Config{
		User: os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASSWORD"),
		Net: "tcp",
		Addr: "127.0.0.1:3306",
		DBName: "todo",
	}
	var err error
	db , err = sql.Open("mysql", config.FormatDSN())

	if err != nil {
		return err
	}

	pingError := db.Ping()

	if pingError != nil {
		return err
	}

	log.Print("Successfully connected to the db.")
	return nil
}

func getUsers(ctx echo.Context) error {
	users, err := db.Query("SELECT * FROM user")

	if err != nil {
		fmt.Println(fmt.Sprintf("**** DB ERROR **** , %s", err))
		return ctx.JSON(http.StatusInternalServerError, "Failed to get users")
	}

	defer users.Close()

	response := []User{}


	for users.Next() {
		var usr UserSchema
		if err := users.Scan(&usr.ID, &usr.UUID, &usr.FirstName, &usr.LastName, &usr.Age); err != nil {
			log.Print(err)
			return err
		}
		user := usr.mapSchemaToUser()
		response = append(response, user)

	}
	return ctx.JSON(http.StatusOK , response)
}

func getUserById(ctx echo.Context) error {
	id := ctx.Param("id")
	var user UserSchema

	row := db.QueryRow("SELECT * FROM USER WHERE uuid = ?", id)

	if err := row.Scan(&user.ID, &user.UUID, &user.FirstName, &user.LastName, &user.Age); err != nil {
		log.Print(err)
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to get user with id: %s", id))
	}

	response := user.mapSchemaToUser()

	return ctx.JSON(http.StatusOK, response)
}

func (user *UserSchema) mapSchemaToUser() User {
	return User {
		ID:        user.UUID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Age: int(user.Age),
	}
}

func (user *User) mapUserToSchema() UserSchema {
	return UserSchema{
		FirstName: user.FirstName,
		LastName: user.LastName,
		Age: int64(user.Age),
		UUID: uuid.New().String(),
	}
}

func createUser(ctx echo.Context) error {
	reqBody := User{}
	defer ctx.Request().Body.Close()
	err := json.NewDecoder(ctx.Request().Body).Decode(&reqBody)

	if err != nil {
		log.Print(err)
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}

	user := reqBody.mapUserToSchema()
	log.Print(&user)

	res , err := db.Exec("INSERT INTO user (uuid, first_name, last_name, age) VALUES (?,?,?,?)", user.UUID, user.FirstName, user.LastName, user.Age)

	if err != nil {
		log.Print(err)
		if err := ctx.JSON(http.StatusInternalServerError, "Failed to create new user"); err != nil {
			return err
		}
	}

	log.Print(res)
	return ctx.JSON(http.StatusOK, reqBody)
}

// echo context represents the context of the currect http request
func main() {
	if err := dbConfig(); err != nil {
		log.Fatal(err)
	}


	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/users", getUsers)
	e.POST("/users", createUser)
	e.GET("/user/:id", getUserById)
	e.Logger.Fatal(e.Start(":8080"))
}
