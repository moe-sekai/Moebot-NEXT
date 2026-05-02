package commands

import (
	"fmt"
	"sort"
	"time"

	"moebot-next/internal/assets"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// RegisterBirthday registers the /生日 command.
func RegisterBirthday() {
	zero.OnCommand("生日").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		now := time.Now()

		// Today's birthdays
		todayBirthdays := assets.GetTodayBirthdays(now)
		upcoming := assets.GetUpcomingBirthdays(now, 30) // next 30 days

		var text string
		if len(todayBirthdays) > 0 {
			text = "🎂 今日生日:\n"
			for _, b := range todayBirthdays {
				char := assets.GetCharacterByID(b.CharacterID)
				if char != nil {
					text += fmt.Sprintf("  🎉 %s (%d月%d日)\n", char.NameCN, b.Month, b.Day)
				}
			}
			text += "\n"
		}

		if len(upcoming) > 0 {
			text += "📅 近期角色生日:\n"
			// Sort by upcoming date
			sort.Slice(upcoming, func(i, j int) bool {
				di := daysUntil(now, upcoming[i].Month, upcoming[i].Day)
				dj := daysUntil(now, upcoming[j].Month, upcoming[j].Day)
				return di < dj
			})

			for _, b := range upcoming {
				char := assets.GetCharacterByID(b.CharacterID)
				if char != nil {
					days := daysUntil(now, b.Month, b.Day)
					if days == 0 {
						continue // already shown as today
					}
					text += fmt.Sprintf("  %s — %d月%d日 (还有%d天)\n", char.NameCN, b.Month, b.Day, days)
				}
			}
		} else if len(todayBirthdays) == 0 {
			text = "近期没有角色生日~"
		}

		ctx.SendChain(message.Text(text))
	})
}

// daysUntil calculates days from now until the next occurrence of month/day.
func daysUntil(now time.Time, month, day int) int {
	year := now.Year()
	target := time.Date(year, time.Month(month), day, 0, 0, 0, 0, now.Location())
	if target.Before(now) {
		target = time.Date(year+1, time.Month(month), day, 0, 0, 0, 0, now.Location())
	}
	return int(target.Sub(now).Hours() / 24)
}
