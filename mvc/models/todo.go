//declare package "models"
package models

// "time" : time for CreatedAt column
// "mvc/config" : to access to global var config.DB in ../mvc/config/config.go
import (
	"time"
	"mvc/config"
)

// Todo is a data model of DB
// `...` are read by enconding/json, gorm 
// ID uint  = unsigned integer
// Title string = what to do
// Completed bool = whether done or not
// CreatedAt time.Time = when it created in DB
type Todo struct {
	ID uint `json:"id" gorm: "primaryKey"`
	Title string `json:"title" gorm:"not null"`
	Completed bool `json:"completed" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"` // gorm automatically assign
}

// GetAllTodos show every to do in the DB
// declare Todo slice
// config.DB = global var DB in ../mvc/config/config.go
// Find(&todos) = SELECT * FROM todos
// &todos is the addr of todos
// result of query get stored in result var
// finally, todos returned
// return todos becuase todos are now updated by Find function which used addr of todos (&todos)  
// result is just a metadata of GORM execution
func GetAllTodos() ([]Todo, error) {
	
	var todos []Todo
	
	result := config.DB.Find(&todos)
	return todos, result.Error
}

// GetTodoByID gets ID to search certain todo using ID and returns a todo using the ID
// declare empty todo whose type is Todo  
// DB.First(&todo, id) = SELECT * FROM todos WHERE id = ?
// use addr of todo to change todo in "First()"
// &todo means addr of todo 
// if error then return no pointer and error message
// else return addr of todo(changed by First()) and no error 	
func GetTodoByID(id uint) (*Todo, error) {
	var todo Todo 
	result := config.DB.First(&todo, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &todo, nil
}

// CreateTodo creates a todo in DB with a title only.
// 1. make new todo struct with the title input
// 2. DB.Create(&todo): INSERT INTO todos ...
// 3. if err return err
func CreateTodo(title string) error {
	todo := Todo{
		Title: title,
		Completed: false,
	}
	result := config.DB.Create(&todo)
	return result.Error
}

// UpdateTodo updates search and update certain todo using inputs : ID uint, completed bool 
// And returns error if it made
// query : UPDATE todos SET completed = ? WHERE id = ?
func UpdateTodo(id uint, completed bool) error {
	result := config.DB.Model(&Todo{}).Where("id = ?", id).Update("completed", completed)
	return result.Error
}

// DeleteTodo deletes a todo searched by ID input 
// And return error if it made
// Delete(&Todo{}, id): DELETE FROM todos WHERE id = ?
// &Todo{} means, firstly, create empty Todo struct and, secondly, get addr of the created struct(&)
func DeleteTodo(id uint) error {
	result := config.DB.Delete(&Todo{}, id)
	return result.Error
}







