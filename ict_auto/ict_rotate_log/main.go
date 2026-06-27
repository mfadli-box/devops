package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type RetentionConfig struct {
	ArchiveDir    string
	NormalLogDays int
	AttackLogDays int
}

func LoadRetentionConfig() *RetentionConfig {
	dir := os.Getenv("RE_PATH")
	if dir == "" {
		dir = "./archive"
	}

	normalDays, _ := strconv.Atoi(os.Getenv("RE_NORMAL"))
	if normalDays <= 0 {
		normalDays = 7
	}

	attackDays, _ := strconv.Atoi(os.Getenv("RE_ATTACK"))
	if attackDays <= 0 {
		attackDays = 90
	}
	return &RetentionConfig{
		ArchiveDir:    dir,
		NormalLogDays: normalDays,
		AttackLogDays: attackDays,
	}
}

var PgSQL *sql.DB

func StartAutoArchiveAndRetentionWorker(db *sql.DB) {
	cfg := LoadRetentionConfig()

	if err := os.MkdirAll(cfg.ArchiveDir, 0755); err != nil {
		log.Fatalf("Gagal membuat folder arsip log: %v", err)
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			log.Println("Memulai siklus harian arsip dan pembersihan...")
			archiveTable(db, "ict_nginx_log", cfg.ArchiveDir, cfg.NormalLogDays)
			archiveTable(db, "ict_nginx_app", cfg.ArchiveDir, cfg.AttackLogDays)
			archiveTable(db, "ict_nginx_atc", cfg.ArchiveDir, cfg.AttackLogDays)
			<-ticker.C
		}
	}()
	runtime.GC()
	debug.FreeOSMemory()
}

func archiveTable(db *sql.DB, tableName string, archiveDir string, retentionDays int) {
	fileName := fmt.Sprintf("%s_archive_%s.log", tableName, time.Now().Format("2006-01-02"))
	filePath := filepath.Join(archiveDir, fileName)
	selectQuery := fmt.Sprintf(`
		SELECT row_to_json(t) FROM (
			SELECT * FROM %s WHERE created_at < NOW() - $1 * INTERVAL '1 day'
		) t`, tableName)
	rows, err := db.Query(selectQuery, retentionDays)
	if err != nil {
		log.Printf("Gagal mengambil data %s: %v", tableName, err)
		return
	}
	defer rows.Close()

	var file *os.File
	var fileCreated bool
	var count int

	for rows.Next() {
		var jsonRaw string
		if err := rows.Scan(&jsonRaw); err != nil {
			log.Printf("Gagal memparsing baris JSON: %v", err)
			continue
		}
		if !fileCreated {
			file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Printf("Gagal membuat file arsip %s: %v", filePath, err)
				return
			}
			defer file.Close()
			fileCreated = true
		}
		if _, err := file.WriteString(jsonRaw + "\n"); err != nil {
			log.Printf("Gagal menulis ke file: %v", err)
			return
		}
		count++
	}
	if count > 0 {
		log.Printf("Berhasil mencadangkan %d baris data dari %s ke %s", count, tableName, filePath)

		deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE created_at < NOW() - $1 * INTERVAL '1 day'", tableName)
		res, err := db.Exec(deleteQuery, retentionDays)
		if err != nil {
			log.Printf("Gagal menghapus data %s setelah diarsip: %v", tableName, err)
			return
		}

		rowsDel, _ := res.RowsAffected()
		log.Printf("Berhasil menghapus %d baris data lama dari tabel %s", rowsDel, tableName)

		_, _ = db.Exec(fmt.Sprintf("VACUUM %s", tableName))
	} else {
		log.Printf("Tidak ada data lama di tabel %s yang perlu diarsip hari ini.", tableName)
	}
}

func main() {
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

	StartAutoArchiveAndRetentionWorker(PgSQL)

	log.Println("Aplikasi berjalan...")
	select {}
}
