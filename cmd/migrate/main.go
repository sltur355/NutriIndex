// C:\RIP\LAB1\cmd\migrate\main.go
package main

import (
	"LAB1/internal/app/ds"
	"LAB1/internal/app/dsn"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	log.Println("Starting database migration...")

	// Очищаем только проблемные таблицы
	log.Println("Cleaning related tables...")

	// Временное отключение foreign key проверок
	db.Exec("SET session_replication_role = 'replica'")

	// Удаляем только таблицы, связанные с users
	db.Exec("DROP TABLE IF EXISTS research_biomarkers CASCADE")
	db.Exec("DROP TABLE IF EXISTS ini_researches CASCADE")
	db.Exec("DROP TABLE IF EXISTS users CASCADE")

	// Включаем обратно foreign key проверки
	db.Exec("SET session_replication_role = 'origin'")

	// Мигрируем все таблицы заново
	log.Println("Creating tables...")
	err = db.AutoMigrate(
		&ds.User{},
		&ds.Biomarker{},
		&ds.INIResearch{},
		&ds.ResearchBiomarker{},
	)
	if err != nil {
		log.Printf("Migration error: %v", err)
		panic("cant migrate db")
	}

	log.Println("Migration completed successfully!")
	log.Println("You can now register new users.")
}
