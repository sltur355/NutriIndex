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

// GetBiomarkers - GET список биомаркеров с фильтрацией и пагинацией
func (r *INIModel) GetBiomarkers(name string, page, limit int, useIndex bool) ([]ds.Biomarker, int64, time.Duration, error) {
	start := time.Now()

	var count int64
	var biomarkers []ds.Biomarker

	// Для демонстрации разницы используем разные подходы
	if !useIndex {
		// Метод БЕЗ использования индекса - принудительный seq scan
		// Используем LOWER() вместо ILIKE чтобы усложнить поиск
		if name != "" {
			// Подсчет
			err := r.db.Raw("SELECT COUNT(*) FROM biomarkers WHERE LOWER(name) LIKE LOWER(?)", "%"+name+"%").Scan(&count).Error
			if err != nil {
				return nil, 0, 0, err
			}

			// Основной запрос
			offset := (page - 1) * limit
			err = r.db.Raw(`
				SELECT * FROM biomarkers 
				WHERE LOWER(name) LIKE LOWER(?) 
				ORDER BY id 
				LIMIT ? OFFSET ?
			`, "%"+name+"%", limit, offset).Scan(&biomarkers).Error

			executionTime := time.Since(start)
			return biomarkers, count, executionTime, err
		} else {
			err := r.db.Model(&ds.Biomarker{}).Count(&count).Error
			if err != nil {
				return nil, 0, 0, err
			}

			offset := (page - 1) * limit
			err = r.db.Offset(offset).Limit(limit).Order("id").Find(&biomarkers).Error

			executionTime := time.Since(start)
			return biomarkers, count, executionTime, err
		}
	}

	// Метод С использованием индекса (оригинальный код)
	// 1. Подсчет общего количества
	countQuery := r.db.Model(&ds.Biomarker{})
	if name != "" {
		countQuery = countQuery.Where("name ILIKE ?", "%"+name+"%")
	}

	err := countQuery.Count(&count).Error
	if err != nil {
		return nil, 0, 0, err
	}

	// 2. Основной запрос с пагинацией
	query := r.db.Model(&ds.Biomarker{})

	// Применяем пагинацию
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	// Сортировка для консистентности пагинации
	query = query.Order("id")

	err = query.Find(&biomarkers).Error
	executionTime := time.Since(start)

	return biomarkers, count, executionTime, err
}

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

// GetQueryPlan - получить план выполнения запроса для демонстрации
func (r *INIModel) GetQueryPlan(name string, useIndex bool) (string, error) {
	var plan string

	sql := "EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) SELECT * FROM biomarkers"
	params := []interface{}{}

	if name != "" {
		if !useIndex {
			// Без индекса
			sql += " WHERE LOWER(name) LIKE LOWER(?)"
		} else {
			// С индексом
			sql += " WHERE name ILIKE ?"
		}
		params = append(params, "%"+name+"%")
	}

	sql += " ORDER BY id LIMIT 20"

	err := r.db.Raw(sql, params...).Scan(&plan).Error
	return plan, err
}
