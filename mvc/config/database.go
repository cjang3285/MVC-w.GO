package config

import (
    "fmt" // for ouput
    "log" // for logging
    "gorm.io/driver/postgres" // PostgreSQL driver
    "gorm.io/gorm" // ORM lib
)

// global DB var declaration. 
// can access DB from other packages using config.DB 
// init is Capital, hence, public <- accessible from outside
var DB *gorm.DB 

func ConnectDB() error {
	var err error
	//DSN : data source name = DB connection information string
	dsn := "host=localhost user=myuser password=mypassword dbname=db port=5432 sslmode=disable"	
	
	// gorm.Open():Connect to DB using GORM
	// postgres.Open(dsn): using Psql driver
	// &gorm.Config{} : GORM basic configuration 
	// result be stored in global var DB,
	
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	
	if err != nil {
		return fmt.Errorf("DB connection failed: %w", err)
	}
		
	fmt.Println("DB connection successful!")
	
	//return type is error, hence, returning nil means no error in the function
	return nil
}

//function that Terminates the Connection to DB
func CloseDB(){
	//check DB is initialized
	if DB != nil {
		// bring *sql.DB from *gorm.DB to terminate the connection
		sqlDB, err := DB.DB()
		// if no error
		if err == nil {
			//terminate the connection
			sqlDB.Close()
			log.Println("DB connection terminated")	
		}
	}
}
