package repository

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type INIModel struct {
	db        *gorm.DB
	minio     *minio.Client
	bucket    string
	jwtSecret string
}

func NewINIModel(dsn string, minioEndpoint, minioAccessKey, minioSecretKey, bucket string, useSSL bool) (*INIModel, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Инициализация MinIO клиента
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %v", err)
	}

	// Проверяем существование бакета
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	// Получаем JWT секрет из переменных окружения
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "fallback-secret-key-change-in-production"
	}

	return &INIModel{
		db:        db,
		minio:     minioClient,
		bucket:    bucket,
		jwtSecret: jwtSecret,
	}, nil
}

// GetJWTSecret возвращает секрет для JWT
func (r *INIModel) GetJWTSecret() string {
	return r.jwtSecret
}

// Фиксированные пользователи как в задании
func GetFixedCreatorID() uint {
	return 4
}

func GetFixedModeratorID() uint {
	return 3
}

// MinIO методы
func (r *INIModel) UploadFileToMinIO(ctx context.Context, fileName string, fileReader io.Reader, fileSize int64, contentType string) error {
	_, err := r.minio.PutObject(ctx, r.bucket, fileName, fileReader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (r *INIModel) DeleteFileFromMinIO(ctx context.Context, fileName string) error {
	err := r.minio.RemoveObject(ctx, r.bucket, fileName, minio.RemoveObjectOptions{})
	return err
}
