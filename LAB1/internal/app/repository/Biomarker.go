package repository

import (
	"fmt"

	"LAB1/internal/app/ds"
)

func (r *INIModel) GetBiomarkers() ([]ds.Biomarker, error) {
	var biomarkers []ds.Biomarker
	err := r.db.Find(&biomarkers).Error
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	if err != nil {
		return nil, err
	}
	if len(biomarkers) == 0 {
		return nil, fmt.Errorf("массив биомаркеров пустой")
	}

	return biomarkers, nil
}

// GetBiomarker возвращает биомаркер по ID
func (r *INIModel) GetDetailedBiomarker(id int) (ds.Biomarker, error) {
	var biomarker = ds.Biomarker{}
	err := r.db.Where("id = ?", id).First(&biomarker).Error
	if err != nil {
		return ds.Biomarker{}, err
	}
	return biomarker, nil
}

func (r *INIModel) GetBiomarkersByName(name string) ([]ds.Biomarker, error) {
	var biomarkers []ds.Biomarker
	err := r.db.Where("name ILIKE ?", "%"+name+"%").Find(&biomarkers).Error
	if err != nil {
		return nil, err
	}
	return biomarkers, nil
}
