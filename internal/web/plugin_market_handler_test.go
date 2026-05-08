package web

import "testing"

// 截取 ZeroBot-Plugin README 的一小段真实片段用于回归：
// - 覆盖 H3 分组（优先级）切换
// - 覆盖 summary / import / 命令行 / 描述行
// - 覆盖 import 路径基名与 <summary> 目录名不一致的容错（aifalse vs ai_false）
const readmeFixture = "## 功能\n" +
	"### *高优先级*\n" +
	"<details>\n" +
	"  <summary>签到</summary>\n\n" +
	"  `import _ \"github.com/FloatTech/ZeroBot-Plugin/plugin/fortune\"`\n\n" +
	"  - [x] 运势\n  - [x] 抽签\n" +
	"</details>\n" +
	"### *中优先级*\n" +
	"<details>\n" +
	"  <summary>AIfalse</summary>\n" +
	"  `import _ \"github.com/FloatTech/ZeroBot-Plugin/plugin/ai_false\"`\n" +
	"  基于活跃度判断的自检插件。\n" +
	"  - [x] 检查身体\n" +
	"</details>\n" +
	"<details>\n" +
	"  <summary>定时指令触发器</summary>\n" +
	"  `import _ \"github.com/FloatTech/zbputils/job\"`\n" +
	"  - [x] 记录指令\n" +
	"</details>\n"

func TestParseReadmeExtractsMetadata(t *testing.T) {
	entries := parseReadme(readmeFixture)
	if len(entries) != 3 {
		t.Fatalf("want 3 entries, got %d", len(entries))
	}

	wantByName := map[string]readmeEntry{
		"fortune": {
			Title:      "签到",
			Priority:   "high",
			Source:     "zerobot-plugin",
			ImportPath: "github.com/FloatTech/ZeroBot-Plugin/plugin/fortune",
		},
		"ai_false": {
			Title:       "AIfalse",
			Priority:    "medium",
			Source:      "zerobot-plugin",
			ImportPath:  "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_false",
			Description: "基于活跃度判断的自检插件。",
		},
		"job": {
			Title:      "定时指令触发器",
			Priority:   "medium",
			Source:     "zbputils",
			ImportPath: "github.com/FloatTech/zbputils/job",
		},
	}

	got := make(map[string]readmeEntry, len(entries))
	for _, e := range entries {
		got[e.Name] = e
	}
	for name, want := range wantByName {
		e, ok := got[name]
		if !ok {
			t.Errorf("missing entry %q", name)
			continue
		}
		if e.Title != want.Title {
			t.Errorf("%s title = %q want %q", name, e.Title, want.Title)
		}
		if e.Priority != want.Priority {
			t.Errorf("%s priority = %q want %q", name, e.Priority, want.Priority)
		}
		if e.Source != want.Source {
			t.Errorf("%s source = %q want %q", name, e.Source, want.Source)
		}
		if e.ImportPath != want.ImportPath {
			t.Errorf("%s import = %q want %q", name, e.ImportPath, want.ImportPath)
		}
		if want.Description != "" && e.Description != want.Description {
			t.Errorf("%s desc = %q want %q", name, e.Description, want.Description)
		}
	}

	// 命令提取回归：fortune 应有两条命令。
	if n := len(got["fortune"].Commands); n != 2 {
		t.Errorf("fortune commands = %d want 2 (%v)", n, got["fortune"].Commands)
	}

	// normalizeName 匹配容错：aifalse 与 ai_false 应命中同一条目。
	if normalizeName("aifalse") != normalizeName("ai_false") {
		t.Errorf("normalizeName collapse failed")
	}
}
