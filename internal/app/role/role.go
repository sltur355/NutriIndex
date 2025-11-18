package role

type Role int

const (
	Guest   Role = iota // 0 - Гость (неавторизованный)
	Patient             // 1 - Пациент (просмотр своих исследований)
	Doctor              // 2 - Врач (модератор)
)

func (r Role) String() string {
	switch r {
	case Doctor:
		return "doctor"
	case Patient:
		return "patient"
	default:
		return "guest"
	}
}

func FromString(s string) Role {
	switch s {
	case "doctor", "Doctor", "moderator", "Moderator":
		return Doctor
	case "patient", "Patient", "user", "User":
		return Patient
	default:
		return Guest
	}
}

// HasPermission проверяет имеет ли роль достаточные права
func (r Role) HasPermission(required Role) bool {
	return r >= required
}

// CanViewResearches может ли просматривать исследования
func (r Role) CanViewResearches() bool {
	return r >= Patient
}

// CanViewAllResearches может ли просматривать ВСЕ исследования
func (r Role) CanViewAllResearches() bool {
	return r >= Doctor
}

// CanCreateResearch может ли создавать исследования
func (r Role) CanCreateResearch() bool {
	return r >= Patient // Пациенты могут создавать исследования!
}

// CanManageResearches может ли управлять исследованиями (завершать/отклонять)
func (r Role) CanManageResearches() bool {
	return r >= Doctor // Только врачи-модераторы
}

// CanManageBiomarkers может ли управлять биомаркерами
func (r Role) CanManageBiomarkers() bool {
	return r >= Doctor // Врачи могут управлять биомаркерами
}

// CanManageUsers может ли управлять пользователями
func (r Role) CanManageUsers() bool {
	return false // В этой системе нет управления пользователями
}
