package assets

import (
	"sort"
	"time"
)

// Birthday associates a character with their birthday month and day.
type Birthday struct {
	CharacterID int
	Month       int // 1–12
	Day         int // 1–31
}

// birthdayData is the hard-coded birthday table for all 26 characters
// (virtual singers 21-26 included).
//
// Source key: charID → month/day
//
//	3→10/27, 25→11/5, 11→11/12, 8→12/6, 22→12/27, 23→12/27,
//	4→1/8,  18→1/27, 24→1/30,  17→2/10, 26→2/17,  9→3/2,
//	7→3/19,  5→4/14, 19→4/30,   2→5/9, 13→5/17,  12→5/25,
//	16→6/24, 15→7/20, 10→7/26,  1→8/11, 20→8/27,  21→8/31,
//	14→9/9,   6→10/5
var birthdayData = []Birthday{
	{1, 8, 11},
	{2, 5, 9},
	{3, 10, 27},
	{4, 1, 8},
	{5, 4, 14},
	{6, 10, 5},
	{7, 3, 19},
	{8, 12, 6},
	{9, 3, 2},
	{10, 7, 26},
	{11, 11, 12},
	{12, 5, 25},
	{13, 5, 17},
	{14, 9, 9},
	{15, 7, 20},
	{16, 6, 24},
	{17, 2, 10},
	{18, 1, 27},
	{19, 4, 30},
	{20, 8, 27},
	{21, 8, 31},
	{22, 12, 27},
	{23, 12, 27},
	{24, 1, 30},
	{25, 11, 5},
	{26, 2, 17},
}

// birthdayMap is a fast lookup keyed by character ID.
var birthdayMap map[int]Birthday

func init() {
	birthdayMap = make(map[int]Birthday, len(birthdayData))
	for _, b := range birthdayData {
		birthdayMap[b.CharacterID] = b
	}
}

// GetBirthday returns the birthday for a given character ID, and whether it
// was found.
func GetBirthday(charID int) (Birthday, bool) {
	b, ok := birthdayMap[charID]
	return b, ok
}

// GetTodayBirthdays returns all characters whose birthday matches the month
// and day of the supplied time.
func GetTodayBirthdays(now time.Time) []Birthday {
	m, d := int(now.Month()), now.Day()
	var result []Birthday
	for _, b := range birthdayData {
		if b.Month == m && b.Day == d {
			result = append(result, b)
		}
	}
	return result
}

// GetUpcomingBirthdays returns all characters whose birthday falls within the
// next `days` calendar days (inclusive of today). Results are sorted by how
// soon they occur.
func GetUpcomingBirthdays(now time.Time, days int) []Birthday {
	if days <= 0 {
		return nil
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	type ranked struct {
		birthday Birthday
		daysAway int
	}

	var hits []ranked
	for _, b := range birthdayData {
		// Build a candidate date in the current year.
		candidate := time.Date(today.Year(), time.Month(b.Month), b.Day, 0, 0, 0, 0, today.Location())
		if candidate.Before(today) {
			// Already passed this year – wrap to next year.
			candidate = candidate.AddDate(1, 0, 0)
		}
		diff := int(candidate.Sub(today).Hours() / 24)
		if diff <= days {
			hits = append(hits, ranked{b, diff})
		}
	}

	sort.Slice(hits, func(i, j int) bool {
		return hits[i].daysAway < hits[j].daysAway
	})

	result := make([]Birthday, len(hits))
	for i, h := range hits {
		result[i] = h.birthday
	}
	return result
}

// IsVirtualSinger reports whether a character ID corresponds to one of the
// six virtual singers (Miku, Rin, Len, Luka, MEIKO, KAITO).
func IsVirtualSinger(charID int) bool {
	return charID >= 21 && charID <= 26
}
