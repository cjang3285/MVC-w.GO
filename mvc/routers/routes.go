package routes


// mvc/controllers : handler functions
// github.com/gorilla/mux : router lib
import (
    "mvc/controllers"
    "github.com/gorilla/mux"
)

// SetupRoutes 
// create new router r
// r.HandleFunc("...", controllers.func) connects the function to call for specified URL or URLs 
// even if URLs same, the function to call can be different by a method 
func SetupRoutes() *mux.Router {
    r := mux.NewRouter()
    
    r.HandleFunc("/todos", controllers.GetTodos).Methods("GET")
    r.HandleFunc("/todos", controllers.CreateTodo).Methods("POST")
    r.HandleFunc("/todos/{id}", controllers.GetTodo).Methods("GET")
    r.HandleFunc("/todos/{id}", controllers.UpdateTodo).Methods("PUT")
    r.HandleFunc("/todos/{id}", controllers.DeleteTodo).Methods("DELETE")
    
    return r
}
