package autochat

import "testing"

func TestParseAutoChatXMLPlainResponse(t *testing.T) {
	raw := `<response>
  <replies>
    <reply>你好呀！</reply>
  </replies>
  <dialogue_summary>打招呼</dialogue_summary>
  <update_profiles></update_profiles>
  <add_memories></add_memories>
</response>`

	got, clean, err := parseAutoChatXML(raw)
	if err != nil {
		t.Fatalf("parseAutoChatXML() error = %v; clean=%q", err, clean)
	}
	if len(got.Replies) != 1 || got.Replies[0] != "你好呀！" {
		t.Fatalf("replies = %#v, want one greeting", got.Replies)
	}
	if got.DialogueSummary != "打招呼" {
		t.Fatalf("dialogue summary = %q, want 打招呼", got.DialogueSummary)
	}
}

func TestParseAutoChatXMLPrefersFinalResponseAfterThinkingDraft(t *testing.T) {
	raw := `*   User: 東雪 (1768184865)
    *   Message: "mmj有哪些队员"

    *   Response 1: List the members enthusiastically.
    *   "MMJ当然是我、遥前辈、爱理酱和雫酱啦！✨"

    *   ` + "`<response>`" + `
    *   ` + "`<replies>`" + `
    *   ` + "`<reply>MMJ当然是我、遥前辈、爱理酱和雫酱啦！✨</reply>`" + `
    *   ` + "`<dialogue_summary>東雪询问MMJ的成员。</dialogue_summary>`" + `
` + "`" + `
<response>
  <replies>
    <reply>MMJ当然是我、遥前辈、爱理酱和雫酱啦！✨</reply>
    <reply>我们四个人会一起努力成为最棒的偶像的！🌟</reply>
  </replies>
  <dialogue_summary>東雪询问MMJ的成员，实乃理热情地介绍了组合成员并表达了决心。</dialogue_summary>
  <update_profiles></update_profiles>
  <add_memories></add_memories>
</response>`

	got, clean, err := parseAutoChatXML(raw)
	if err != nil {
		t.Fatalf("parseAutoChatXML() error = %v; clean=%q", err, clean)
	}
	if len(got.Replies) != 2 {
		t.Fatalf("replies len = %d, want 2: %#v", len(got.Replies), got.Replies)
	}
	if got.Replies[0] != "MMJ当然是我、遥前辈、爱理酱和雫酱啦！✨" {
		t.Fatalf("first reply = %q", got.Replies[0])
	}
	if got.Replies[1] != "我们四个人会一起努力成为最棒的偶像的！🌟" {
		t.Fatalf("second reply = %q", got.Replies[1])
	}
	if clean[:len("<response>")] != "<response>" {
		t.Fatalf("clean should start with final response, got %q", clean)
	}
}

func TestExtractXMLBlockPrefersLastCompleteResponse(t *testing.T) {
	raw := `<response>
  <replies>
    <reply>草稿</reply>
  </replies>
</response>
noise
<response>
  <replies>
    <reply>最终</reply>
  </replies>
  <dialogue_summary></dialogue_summary>
  <update_profiles></update_profiles>
  <add_memories></add_memories>
</response>`

	clean := extractXMLBlock(raw)
	want := `<response>
  <replies>
    <reply>最终</reply>
  </replies>
  <dialogue_summary></dialogue_summary>
  <update_profiles></update_profiles>
  <add_memories></add_memories>
</response>`
	if clean != want {
		t.Fatalf("clean = %q, want %q", clean, want)
	}
}

func TestParseAutoChatXMLInsideMarkdownFence(t *testing.T) {
	raw := "```xml\n<response>\n  <replies>\n    <reply>围栏里也可以</reply>\n  </replies>\n  <dialogue_summary></dialogue_summary>\n  <update_profiles></update_profiles>\n  <add_memories></add_memories>\n</response>\n```"

	got, clean, err := parseAutoChatXML(raw)
	if err != nil {
		t.Fatalf("parseAutoChatXML() error = %v; clean=%q", err, clean)
	}
	if len(got.Replies) != 1 || got.Replies[0] != "围栏里也可以" {
		t.Fatalf("replies = %#v, want fenced reply", got.Replies)
	}
}
