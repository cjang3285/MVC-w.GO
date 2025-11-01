package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    "sync"
)

// Todo 구조체
type Todo struct {
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

// 전역 변수로 todos 저장 (실제로는 데이터베이스 사용)
var (
    todos  = []Todo{}
    nextID = 1
    mu     sync.Mutex // 동시성 처리를 위한 뮤텍스
)

func main() {
    // 초기 데이터
    todos = append(todos, Todo{ID: nextID, Title: "Go 공부하기", Completed: false})
    nextID++
    todos = append(todos, Todo{ID: nextID, Title: "API 만들기", Completed: false})
    nextID++
    
    // 라우트 설정
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/todos", todosHandler)
    http.HandleFunc("/todos/", todoHandler)
    
    fmt.Println("서버 시작: http://localhost:8080")
    fmt.Println("사용 가능한 엔드포인트:")
    fmt.Println("  GET    /todos       - 모든 todo 조회")
    fmt.Println("  POST   /todos       - 새 todo 생성")
    fmt.Println("  GET    /todos/{id}  - 특정 todo 조회")
    fmt.Println("  PUT    /todos/{id}  - todo 수정")
    fmt.Println("  DELETE /todos/{id}  - todo 삭제")
    
    http.ListenAndServe(":8080", nil)
}

// 홈 핸들러
func homeHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    response := map[string]string{
        "message": "Todo API에 오신 것을 환영합니다",
        "version": "1.0",
    }
    
    json.NewEncoder(w).Encode(response)
}

// /todos 핸들러 (목록 조회, 생성)
func todosHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    switch r.Method {
    case "GET":
        getAllTodos(w, r)
    case "POST":
        createTodo(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// /todos/{id} 핸들러 (개별 조회, 수정, 삭제)
func todoHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    // URL에서 ID 추출
    idStr := strings.TrimPrefix(r.URL.Path, "/todos/")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }
    
    switch r.Method {
    case "GET":
        getTodo(w, r, id)
    case "PUT":
        updateTodo(w, r, id)
    case "DELETE":
        deleteTodo(w, r, id)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// GET /todos - 모든 todo 조회
func getAllTodos(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    defer mu.Unlock()
    
    json.NewEncoder(w).Encode(todos)
}

// POST /todos - 새 todo 생성
func createTodo(w http.ResponseWriter, r *http.Request) {
    var newTodo Todo
    
    if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if newTodo.Title == "" {
        http.Error(w, "Title is required", http.StatusBadRequest)
        return
    }
    
    mu.Lock()
    newTodo.ID = nextID
    nextID++
    todos = append(todos, newTodo)
    mu.Unlock()
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(newTodo)
}

// GET /todos/{id} - 특정 todo 조회
func getTodo(w http.ResponseWriter, r *http.Request, id int) {
    mu.Lock()
    defer mu.Unlock()
    
    for _, todo := range todos {
        if todo.ID == id {
            json.NewEncoder(w).Encode(todo)
            return
        }
    }
    
    http.Error(w, "Todo not found", http.StatusNotFound)
}

// PUT /todos/{id} - todo 수정
func updateTodo(w http.ResponseWriter, r *http.Request, id int) {
    var updatedTodo Todo
    
    if err := json.NewDecoder(r.Body).Decode(&updatedTodo); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    mu.Lock()
    defer mu.Unlock()
    
    for i, todo := range todos {
        if todo.ID == id {
            updatedTodo.ID = id
            todos[i] = updatedTodo
            json.NewEncoder(w).Encode(updatedTodo)
            return
        }
    }
    
    http.Error(w, "Todo not found", http.StatusNotFound)
}

// DELETE /todos/{id} - todo 삭제
func deleteTodo(w http.ResponseWriter, r *http.Request, id int) {
    mu.Lock()
    defer mu.Unlock()
    
    for i, todo := range todos {
        if todo.ID == id {
            todos = append(todos[:i], todos[i+1:]...)
            w.WriteHeader(http.StatusNoContent)
            return
        }
    }
    
    http.Error(w, "Todo not found", http.StatusNotFound)
}
