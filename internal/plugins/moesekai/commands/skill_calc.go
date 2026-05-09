package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"moebot-next/internal/renderer"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// skillCalcCommands lists the chat triggers for the multiplier calculator.
var skillCalcCommands = []string{"倍率计算", "倍率", "skill", "skillcalc"}

// skillCalcNumberRe matches positive integers and decimals.
var skillCalcNumberRe = regexp.MustCompile(`\d+(?:\.\d+)?`)

// SkillCalcPayload is the data shape consumed by the skill_calc satori template.
type SkillCalcPayload struct {
	Inputs       []float64 `json:"inputs"`
	ChariotHead  float64   `json:"chariotHead"`
	Internal     float64   `json:"internal"`
	Multiplier   float64   `json:"multiplier"`
	ActualValue  float64   `json:"actualValue"`
	OthersAvg    float64   `json:"othersAvg"`
	Title        string    `json:"title"`
	Subtitle     string    `json:"subtitle"`
	UsageHint    string    `json:"usageHint"`
}

// RegisterSkillCalc registers the /倍率计算 command. The handler accepts five
// numbers (the user's own card value followed by the four other team-mate
// skill values) and renders the calculated multiplier card via Satori.
func RegisterSkillCalc(deps *Deps) {
	for _, name := range skillCalcCommands {
		Engine.OnCommand(name).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			handleSkillCalc(deps, ctx)
		})
	}
}

func handleSkillCalc(deps *Deps, ctx *zero.Ctx) {
	args := commandArgs(ctx)
	matches := skillCalcNumberRe.FindAllString(args, -1)
	if len(matches) < 5 {
		ctx.SendChain(message.Text(
			"用法: /倍率计算 <车头自身> <队友1> <队友2> <队友3> <队友4>\n" +
				"例: /倍率计算 100 80 80 80 80",
		))
		return
	}
	vals := make([]float64, 0, 5)
	for _, s := range matches[:5] {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			ctx.SendChain(message.Text("无法解析的数字: " + s))
			return
		}
		vals = append(vals, v)
	}
	a, b, c, d, e := vals[0], vals[1], vals[2], vals[3], vals[4]
	othersAvg := (b + c + d + e) / 5
	chariotHead := a
	internal := a + b + c + d + e
	multiplier := (a + othersAvg + 100) / 100
	actual := a + othersAvg

	payload := SkillCalcPayload{
		Inputs:      vals,
		ChariotHead: chariotHead,
		Internal:    internal,
		Multiplier:  multiplier,
		ActualValue: actual,
		OthersAvg:   othersAvg,
		Title:       "卡组技能效果计算",
		Subtitle:    fmt.Sprintf("车头 %s · 队友 %s/%s/%s/%s", trimNum(a), trimNum(b), trimNum(c), trimNum(d), trimNum(e)),
		UsageHint:   "倍率 = (车头 + 其余平均/5 + 100) / 100",
	}

	if deps != nil && deps.Renderer != nil && deps.Renderer.Health() {
		if png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "skill_calc", Data: payload}); err == nil {
			ctx.SendChain(message.ImageBytes(png))
			return
		}
	}
	ctx.SendChain(message.Text(formatSkillCalcText(payload)))
}

func formatSkillCalcText(p SkillCalcPayload) string {
	lines := []string{
		"卡组技能效果计算结果",
		strings.Repeat("-", 24),
		fmt.Sprintf("车头     : %s", trimNum(p.ChariotHead)),
		fmt.Sprintf("内部     : %s", trimNum(p.Internal)),
		fmt.Sprintf("倍率     : %.2f", p.Multiplier),
		fmt.Sprintf("技能实际值: %s%%", trimNum(p.ActualValue)),
		strings.Repeat("-", 24),
	}
	return strings.Join(lines, "\n")
}

// trimNum prints a float with up to 1 decimal, dropping trailing ".0".
func trimNum(v float64) string {
	s := strconv.FormatFloat(v, 'f', 1, 64)
	s = strings.TrimSuffix(s, ".0")
	return s
}
