package main

// log for logging
// net/http for http.ListenAndServe(":8080", router)
// mvc/config for ConnectDB() and CloseDB()
// mvc/routes for routes.SetupRoutes()
import (
    "log"
    "net/http"
    "mvc/config"
    "mvc/routers"
)


func main() {
    if err := config.ConnectDB(); err != nil {
        log.Fatal("DB connetion failed:", err)
    }
    // defer : execute this code after main function ends
    defer config.CloseDB()
    
    router := routes.SetupRoutes()
    
    log.Println(`Server started: http://localhost:8080`)
    log.Fatal(http.ListenAndServe(":8080", router))
}

