package ai

import (
	"strings"

	"github.com/cobrich/ai-tutor-bot/internal/entity"
)

func DetectSubject(question string) entity.Subject {
	q := strings.ToLower(question)

	switch {

	// Математика
	case strings.Contains(q, "реши"),
		strings.Contains(q, "уравнение"),
		strings.Contains(q, "неравенство"),
		strings.Contains(q, "процент"),
		strings.Contains(q, "дроб"),
		strings.Contains(q, "матем"):

		return entity.SubjectMath

	// Английский
	case strings.Contains(q, "english"),
		strings.Contains(q, "translate"),
		strings.Contains(q, "переведи"),
		strings.Contains(q, "present simple"),
		strings.Contains(q, "past simple"):

		return entity.SubjectEnglish

	// Химия
	case strings.Contains(q, "химия"),
		strings.Contains(q, "реакция"),
		strings.Contains(q, "молекула"),
		strings.Contains(q, "кислота"):

		return entity.SubjectChemistry

	// Физика
	case strings.Contains(q, "физика"),
		strings.Contains(q, "скорость"),
		strings.Contains(q, "сила"),
		strings.Contains(q, "энергия"):

		return entity.SubjectPhysics

	// История
	case strings.Contains(q, "история"),
		strings.Contains(q, "война"),
		strings.Contains(q, "революция"):

		return entity.SubjectHistory
	}

	return entity.SubjectUnknown
}