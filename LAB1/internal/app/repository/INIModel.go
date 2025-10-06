package repository

import (
	"fmt"
	"strings"
)

type INIModel struct {
}

func NewINIModel() (*INIModel, error) {
	return &INIModel{}, nil
}

// Biomarker описывает биохимический показатель
type Biomarker struct {
	ID              int
	Name            string
	MeasureUnit     string
	Description     string
	FullDescription string
	MinValue        float64
	MaxValue        float64
	Significance    float64
	ImageURL        string
}

// GetBiomarkers возвращает список биомаркеров
func (r *INIModel) GetBiomarkers() ([]Biomarker, error) {
	biomarkers := []Biomarker{
		{
			ID:              1,
			Name:            "Альбумин",
			MeasureUnit:     "г/л",
			Description:     "основной белок плазмы крови, который производится в печени и выполняет множество функций",
			FullDescription: "Альбумин (ALB) – белок плазмы крови, который выполняет большое количество важных функций. Он участвует в процессах метаболизма, переносит по организму ряд химических веществ и прочее. Определение уровня этого белка играет важную роль в оценке состояния печени. ALB составляет примерно 60% от общего количества белка.",
			MinValue:        35,
			MaxValue:        50,
			Significance:    0.4,
			ImageURL:        "http://127.0.0.1:9000/biomarkers/albumin.png",
		},
		{
			ID:              2,
			Name:            "Лимфоциты",
			MeasureUnit:     "*10⁹/л",
			Description:     "клетки иммунной системы, которые защищают организм от инфекций, бактерий и вирусов",
			FullDescription: "Лимфоциты – клетки иммунной системы, которые защищают организм от инфекций, бактерий и вирусов. Они играют важную роль в иммунной системе и помогают бороться с различными заболеваниями. Лимфоциты составляют примерно 20% от общего количества клеток крови.",
			MinValue:        1.2,
			MaxValue:        4.5,
			Significance:    0.3,
			ImageURL:        "http://127.0.0.1:9000/biomarkers/lymphocytes.png",
		},
		{
			ID:              3,
			Name:            "Холестерин",
			MeasureUnit:     "ммоль/л",
			Description:     "вещество, необходимое для нормального функционирования всех клеток организма",
			FullDescription: "Холестерин (CHO) – вещество, необходимое для нормального функционирования всех клеток организма. Он участвует в процессах метаболизма, переносит по организму ряд химических веществ и прочее. Определение уровня этого вещества играет важную роль в оценке состояния печени. CHO составляет примерно 60% от общего количества вещества.",
			MinValue:        3.2,
			MaxValue:        6.2,
			Significance:    0.2,
			ImageURL:        "http://127.0.0.1:9000/biomarkers/cholesterol.png",
		},
		{
			ID:              4,
			Name:            "ИМТ",
			MeasureUnit:     "кг/м²",
			Description:     "величина, позволяющая оценить степень соответствия массы человека и его роста",
			FullDescription: "Индекс массы тела(ИМТ, англ. body mass index (BMI)) — величина, позволяющая оценить степень соответствия массы человека и его роста и тем самым косвенно судить о том, является ли масса недостаточной, нормальной или избыточной. Важен при определении показаний для необходимости лечения.",
			MinValue:        18.5,
			MaxValue:        24.9,
			Significance:    0.05,
			ImageURL:        "http://127.0.0.1:9000/biomarkers/bmi.png",
		},
		{
			ID:              5,
			Name:            "Общий белок",
			MeasureUnit:     "г/л",
			Description:     "суммарное количество всех видов белков, которые циркулируют в сыворотке крови",
			FullDescription: "Общий белок (TP, англ. total protein) – суммарное количество всех видов белков, которые циркулируют в сыворотке крови. Этот показатель играет важную роль в оценке общего состояния организма. TP составляет примерно 60% от общего количества белка.",
			MinValue:        65,
			MaxValue:        85,
			Significance:    0.05,
			ImageURL:        "http://127.0.0.1:9000/biomarkers/mainprotein.png",
		},
	}

	if len(biomarkers) == 0 {
		return nil, fmt.Errorf("массив биомаркеров пустой")
	}

	return biomarkers, nil
}

// GetBiomarker возвращает биомаркер по ID
func (r *INIModel) GetDetailedBiomarker(id int) (Biomarker, error) {
	biomarkers, err := r.GetBiomarkers()
	if err != nil {
		return Biomarker{}, err
	}

	for _, biomarker := range biomarkers {
		if biomarker.ID == id {
			return biomarker, nil
		}
	}
	return Biomarker{}, fmt.Errorf("биомаркер не найден")
}

// GetBiomarkersByName выполняет поиск по имени
func (r *INIModel) GetBiomarkersByName(name string) ([]Biomarker, error) {
	biomarkers, err := r.GetBiomarkers()
	if err != nil {
		return []Biomarker{}, err
	}

	var result []Biomarker
	for _, biomarker := range biomarkers {
		if strings.Contains(strings.ToLower(biomarker.Name), strings.ToLower(name)) {
			result = append(result, biomarker)
		}
	}

	return result, nil
}

// представляет заявку с данными пациента и выбранными биомаркерами
type MedicalTest struct {
	BiomarkerID  int
	PatientValue float64 // Поле М-М - значение пациента
}

type PatientCart struct {
	ID            int
	PatientName   string
	PatientBirth  string
	PatientGender string
	PatientTests  []MedicalTest // Связь М-М
}

// возвращает заявку с предустановленными данными
func (r *INIModel) GetINIresearch(id int) (map[string]interface{}, error) {
	PatientTests := []MedicalTest{
		{BiomarkerID: 1, PatientValue: 35.2},
		{BiomarkerID: 2, PatientValue: 1.8},
		{BiomarkerID: 3, PatientValue: 4.1},
		{BiomarkerID: 5, PatientValue: 19.2},
	}

	var INIresearchItems []map[string]interface{}
	for _, MedicalTest := range PatientTests {
		biomarker, err := r.GetDetailedBiomarker(MedicalTest.BiomarkerID)
		if err != nil {
			continue
		}

		INIresearchItem := map[string]interface{}{
			"BiomarkerID":    MedicalTest.BiomarkerID,
			"BiomarkerImage": biomarker.ImageURL,
			"BiomarkerName":  biomarker.Name,
			"MeasureUnit":    biomarker.MeasureUnit,
			"MinValue":       biomarker.MinValue,
			"MaxValue":       biomarker.MaxValue,
			"Significance":   biomarker.Significance,
			"PatientValue":   MedicalTest.PatientValue, // Поле М-М
		}
		INIresearchItems = append(INIresearchItems, INIresearchItem)
	}

	result := map[string]interface{}{
		"ID":               1,
		"PatientName":      "Турланов Вячеслав Евгеньевич",
		"PatientBirth":     "12.10.2005",
		"PatientGender":    "Мужской",
		"INIResult":        63.7,
		"INIresearchItems": INIresearchItems,
	}

	return result, nil
}

// GetCartItemsCount возвращает количество биомаркеров в заявке
func (r *INIModel) GetINIresearchItemsCount(id int) (int, error) {
	INIresearch, err := r.GetINIresearch(id)
	if err != nil {
		return 0, err
	}

	// Получаем Items из map
	items, ok := INIresearch["INIresearchItems"].([]map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("неверный формат данных INIresearchItems")
	}

	return len(items), nil
}
