package backbone

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var PgSQL *sql.DB

func SetDatabase() {
	godotenv.Load()
	PG_Host := os.Getenv("PG_HOST")
	PG_Port := os.Getenv("PG_PORT")
	PG_User := os.Getenv("PG_USER")
	PG_Pass := os.Getenv("PG_PASS")
	PG_Data := os.Getenv("PG_DATA")
	IS_Pool := os.Getenv("IS_POOL")

	var dsn string
	dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		PG_Host, PG_Port, PG_User, PG_Pass, PG_Data)

	var err error
	PgSQL, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Gagal membuka basis data: %v", err)
	}

	if err = PgSQL.Ping(); err != nil {
		log.Fatalf("Basis data tidak merespon: %v", err)
	}

	if IS_Pool == "true" {
		PgSQL.SetMaxOpenConns(100)
		PgSQL.SetMaxIdleConns(10)
	} else {
		PgSQL.SetMaxOpenConns(50)
		PgSQL.SetMaxIdleConns(25)
	}
}
