package ds

type ResearchBiomarker struct {
	IDResearch   uint     `gorm:"not null;uniqueIndex:idx_research_biomarker" json:"research_id"`
	IDBiomarker  uint     `gorm:"not null;uniqueIndex:idx_research_biomarker" json:"biomarker_id"`
	PatientValue *float64 `gorm:"type:decimal(10,2);default:null" json:"patient_value,omitempty"`

	Research  *INIResearch `gorm:"foreignKey:IDResearch" json:"research,omitempty"`
	Biomarker *Biomarker   `gorm:"foreignKey:IDBiomarker" json:"biomarker,omitempty"`
}
