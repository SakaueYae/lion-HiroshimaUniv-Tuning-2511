package db

import (
	"backend/internal/telemetry"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func InitDBConnection() (*sqlx.DB, error) {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "user:password@tcp(db:4306)/hiroshimauniv2511-db"
	}
	dsn := fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=UTC", dbUrl)
	// log.Printf(dsn)

	driverName := telemetry.WrapSQLDriver("mysql")
	dbConn, err := sqlx.Open(driverName, dsn)
	if err != nil {
		log.Printf("Failed to open database connection: %v", err)
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = dbConn.PingContext(ctx)
	if err != nil {
		dbConn.Close()
		log.Printf("Failed to connect to database: %v", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	// log.Println("Successfully connected to MySQL!")

	// データベース接続プールの最適化設定
	dbConn.SetMaxOpenConns(100)                // 最大オープン接続数を増加（高負荷時の同時接続に対応）
	dbConn.SetMaxIdleConns(25)                 // アイドル接続数を増加（接続再利用の効率化）
	dbConn.SetConnMaxLifetime(5 * time.Minute) // 5分で接続をリサイクル（古いコネクションを定期的にリフレッシュ）
	dbConn.SetConnMaxIdleTime(1 * time.Minute) // アイドル1分で解放（不要なアイドル接続を早期解放）

	return dbConn, nil
}
