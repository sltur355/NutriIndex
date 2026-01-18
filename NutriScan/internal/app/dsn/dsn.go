package dsn

//удаленное подключение к бд
import (
	"fmt"
	"os"
)

func FromEnv() string {
	host := os.Getenv("DB_HOST")
	if host == "" {
		return ""
	}
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	fmt.Printf("Connecting to: host=%s port=%s user=%s dbname=%s\n", host, port, user, dbname)
	// И вот мы возвращаем dsn, который необходим для подключения к БД
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
}
