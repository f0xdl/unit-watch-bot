package templates

import (
	"fmt"
	"time"
)

type LangCode string

func LangCodeIsValid(lang string) bool {
	switch lang {
	case string(English), string(Ukraine), string(Russian):
		return true
	default:
		return false
	}
}

const (
	English LangCode = "en"
	Ukraine LangCode = "uk"
	Russian LangCode = "ru"
)

// FormatChangedAt format by LangCode
func FormatChangedAt(t time.Time, lang LangCode) string {
	months := map[LangCode][]string{
		"en": {"January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December"},
		"ru": {"января", "февраля", "марта", "апреля", "мая", "июня",
			"июля", "августа", "сентября", "октября", "ноября", "декабря"},
		"uk": {"січня", "лютого", "березня", "квітня", "травня", "червня",
			"липня", "серпня", "вересня", "жовтня", "листопада", "грудня"},
	}

	monthNames, ok := months[lang]
	if !ok {
		panic("invalid lang" + string(lang))
	}

	day := t.Day()
	month := monthNames[int(t.Month())-1]
	year := t.Year()
	hour := t.Hour()
	m := t.Minute()

	return fmt.Sprintf("%02d %s %d, %02d:%02d", day, month, year, hour, m)
}

func FormatSeenAgo(t time.Time, lang LangCode) string {
	delta := time.Since(t)
	var templates []string
	switch lang {
	case English:
		templates = []string{"less than a minute ago", "%d minutes ago", "%d hours ago"}
	case Russian:
		templates = []string{"меньше минуты назад", "%d мин назад", "%d ч назад"}
	case Ukraine:
		templates = []string{"менш ніж хвилину тому", "%d хв тому", "%d год тому"}
	default:
		panic("invalid lang" + string(lang))

	}
	switch {
	case delta < time.Minute:
		return templates[0]
	case delta < time.Hour:
		minutes := int(delta.Minutes())
		return fmt.Sprintf(templates[1], minutes)
	case delta < 24*time.Hour:
		hours := int(delta.Hours())
		return fmt.Sprintf(templates[2], hours)
	default:
		return t.Format("02.01.2006 в 15:04")
	}
}
