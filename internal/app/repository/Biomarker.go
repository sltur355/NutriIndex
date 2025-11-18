package repository

import (
	"LAB1/internal/app/ds"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// GetBiomarkers - GET список биомаркеров с фильтрацией
func (r *INIModel) GetBiomarkers(name string) ([]ds.Biomarker, error) {
	var biomarkers []ds.Biomarker
	query := r.db

	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	err := query.Find(&biomarkers).Error
	return biomarkers, err
}

// GetBiomarker - GET один биомаркер по ID
func (r *INIModel) GetBiomarker(id uint) (ds.Biomarker, error) {
	var biomarker ds.Biomarker
	err := r.db.Where("id = ?", id).First(&biomarker).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Biomarker{}, errors.New("biomarker not found")
		}
		return ds.Biomarker{}, err
	}
	return biomarker, nil
}

// CreateBiomarker - POST создание биомаркера
func (r *INIModel) CreateBiomarker(biomarker ds.Biomarker) (ds.Biomarker, error) {
	biomarker.ID = 0 // гарантируем создание новой записи
	biomarker.CreatedAt = time.Now()
	biomarker.UpdatedAt = time.Now()

	err := r.db.Create(&biomarker).Error
	if err != nil {
		return ds.Biomarker{}, err
	}
	return biomarker, nil
}

// UpdateBiomarker - PUT обновление биомаркера
func (r *INIModel) UpdateBiomarker(id uint, biomarker ds.Biomarker) error {
	// ИЗМЕНЕНИЕ: Обновляем UpdatedAt
	biomarker.UpdatedAt = time.Now()

	tx := r.db.Model(&ds.Biomarker{}).Where("id = ?", id).Updates(biomarker)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("biomarker not found")
	}
	return nil
}

// DeleteBiomarker - DELETE удаление биомаркера и его изображения
func (r *INIModel) DeleteBiomarker(id uint) error {
	// Получаем биомаркер чтобы узнать путь к изображению
	var biomarker ds.Biomarker
	err := r.db.Where("id = ?", id).First(&biomarker).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("biomarker not found")
		}
		return err
	}

	// Удаляем изображение из MinIO если оно есть
	// ИЗМЕНЕНИЕ: Проверяем указатель вместо sql.NullString
	if biomarker.ImageURL != nil && *biomarker.ImageURL != "" {
		fileName := extractFileNameFromURL(*biomarker.ImageURL)
		if fileName != "" {
			ctx := context.Background()
			err = r.DeleteFileFromMinIO(ctx, fileName)
			if err != nil {
				// Логируем ошибку но не прерываем удаление записи
				fmt.Printf("Warning: failed to delete image from MinIO: %v\n", err)
			}
		}
	}

	// Удаляем биомаркер из БД
	tx := r.db.Delete(&ds.Biomarker{}, id)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("biomarker not found")
	}
	return nil
}

// UpdateBiomarkerImage - обновление пути к изображению
func (r *INIModel) UpdateBiomarkerImage(id uint, imagePath string) error {
	// ИЗМЕНЕНИЕ: Используем указатель на строку
	imageURL := &imagePath
	tx := r.db.Model(&ds.Biomarker{}).Where("id = ?", id).Updates(map[string]interface{}{
		"image_url":  imageURL,
		"updated_at": time.Now(),
	})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("biomarker not found")
	}
	return nil
}

// GenerateImageFileName - генерация имени файла на латинице
func (r *INIModel) GenerateImageFileName(originalName string) string {
	ext := ""
	if idx := strings.LastIndex(originalName, "."); idx != -1 {
		ext = originalName[idx:]
	}

	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("biomarker_%d%s", timestamp, ext)
}

// Вспомогательная функция для извлечения имени файла из URL
func extractFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
