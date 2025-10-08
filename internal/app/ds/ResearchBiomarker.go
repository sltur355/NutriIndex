package ds

//м-м
type ResearchBiomarker struct {
	IDResearch   uint    `gorm:"not null;uniqueIndex:idx_research_biomarker"`
	IDBiomarker  uint    `gorm:"not null;uniqueIndex:idx_research_biomarker"`
	PatientValue float64 `gorm:"type:decimal(10,2);not null"`

	// Связи (явные, как в примере)
	Research  INIResearch `gorm:"foreignKey:IDResearch"`
	Biomarker Biomarker   `gorm:"foreignKey:IDBiomarker"`
}
