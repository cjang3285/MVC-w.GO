package controllers

// encoding/json for JSON converting
// net/http for HTTP process
// strconv for string from/to number converting
// mvc/modles for using models package
// github.com/gorilla/mux for elicitation of URL parameters
import (
    "encoding/json"
    "net/http"
    "strconv"
    "mvc/models"
    "github.com/gorilla/mux"
)

// HTTP response structure
// 1. "status line" consist of protocol version, status code, status text (e.g. HTTP/1.1, 200, OK)
// 2. "headers" consist of "content type", content length, authorization, set-cookie ...
// 3. "body" is real data, its format depends on content-type (e.g. JSON, HTML, XML, IMG, ...) 

// HTTP request structure
// 1. "request line" consist of method(GET, POST, PUT, DELETE), URL, protocol version>
// 2. "headers" consist of "Host"(which server domain to send), "Accept"(desired response format from client), content type, content length, authorization, ...
// 3. "body" is real data, its format depends on content-type 


// w.Header().Set() -> Set any header (e.g. content type, authorization, ...) 
// w.WriteHeader() -> set the status line and headers and transmit them (e.g. 201, 404, ...
// json.NewEncoder(w).Encode(data) -> transmit response with body(data)

// all at once function below
// http.Error(w, body, code) : set headers, transmit status code + headers
// and transmit body.  

// Status line, header transmitted, body transmitted.
// they get transmitted one by one.

// GetTodos is a handler to use GetAllTodos() 
// http.Error() : transmit Error response ()
// Header().Set() : set header to notice that this is JSON
// json.NewEncoder(w) : create JSON encoder to w 
// Encode(todos) : to convert todos to JSON and write it to w  
func GetTodos(w http.ResponseWriter, r *http.Request) {
    todos, err := models.GetAllTodos()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todos)
}

// GetTodo is to use models.GetTodoByID(ID) func
// mux.Vars(r) : extract URL parameters, in this case, only ID
// strconv.Atoi : id string(ASCII) to id integer
// if error, wrong ID
// models.GetTodoByID(uint(id)) to search the todo in DB and return it
// if error, it is problem of internal server(func of models)
// set header of w for responding
// encode todo to JSON format to deliver 
func GetTodo(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Wrong ID", http.StatusBadRequest)
        return
    }
    
    todo, err := models.GetTodoByID(uint(id))
    if err != nil a{
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todo)
}


// CreateTodo is to use models.CreateTodo() to insert todo into DB
// make req struct
// To decode r.Body(JSON format), we need a struct to handle the result of conversion
// Which is req struct we made
// e.g. json.NewDecoder(JSON).Decode(req struct) make req JSON -> req struct 
// if err, then request(JSON format) is wrong
// if title is empty, then also wrong
// if models.CreateTodo() fail, then internal problem.
// if no problem, writeheader() set state code that created successfully, and encode success message string to JSON to respond 
func CreateTodo(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Title string `json:"title"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Wrong request", http.StatusBadRequest)
        return
    }
    
    if req.Title == "" {
        http.Error(w, "Title is empty and must be written", http.StatusBadRequest)
        return
    }
    
    if err := models.CreateTodo(req.Title); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Created successfully"})
}

// UpdateTodo to update certain todo is completed or not in DB
// extract ID from request URL
// if converting string(ASCII) ID to integer failed, wrong ID error
// if decoding from JSON request to req struct failed, wrong request error
// if models.UpdateTodo(id, completed) is failed, internal server error  
// if no problem, write response and encode the message map to JSON
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Wrong ID", http.StatusBadRequest)
        return
    }
    
    var req struct {
        Completed bool `json:"completed"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Wrong request", http.StatusBadRequest)
        return
    }
    
    if err := models.UpdateTodo(uint(id), req.Completed); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{"message": "수정 완료"})
}

// DeleteTodo is to call models.DeleteTodo(id) to delete todo from DB
// mux.Vars(r) to extract URL parameter
// pretty much same with others...
// if making id ASCII to Integer Atoi failed, send error "wrong id"
// if models.DeleteTodo(id) failed, send error "internal server error" 
// if no problem, writeheader() sends 204 No Content status
func DeleteTodo(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Wrong ID", http.StatusBadRequest)
        return
    }
    
    if err := models.DeleteTodo(uint(id)); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}
