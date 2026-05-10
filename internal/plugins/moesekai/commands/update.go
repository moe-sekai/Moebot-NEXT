package commands

import (
	"fmt"
	"strings"

	"moebot-next/internal/plugins/moesekai/servers"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// RegisterUpdate registers /update — a SuperUser-only command which forces a
// masterdata refresh on every enabled region. Each region reuses the periodic
// refresh logic: probe versions/current_version.json (camelCase MoeSekai/Haruki)
// or versions.json (snake_case 8823 / Sekai-World), compare with the locally
// persisted dataVersion, and only re-download all master files when the version
// actually changed. Use /update! 强制忽略版本一致性检查执行全量重拉。
func RegisterUpdate(deps *Deps) {
	Engine.OnFullMatchGroup(
		[]string{"/update", "/更新", "/update!", "/强制更新"},
		zero.SuperUserPermission,
	).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		handleUpdate(deps, ctx)
	})
}

func handleUpdate(deps *Deps, ctx *zero.Ctx) {
	if deps == nil || deps.Servers == nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("Masterdata 管理器未就绪"))
		return
	}

	raw := strings.TrimSpace(ctx.Event.RawMessage)
	force := raw == "/update!" || raw == "/强制更新"

	if force {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("正在强制重新拉取 masterdata（忽略版本检查）…"))
	} else {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("正在检查 masterdata 更新…"))
	}

	results := deps.Servers.RefreshAll(force)
	if len(results) == 0 {
		ctx.SendChain(message.Text("当前没有启用的区服"))
		return
	}

	var b strings.Builder
	b.WriteString("Masterdata 更新结果：\n")
	for _, r := range results {
		b.WriteString(formatRefreshLine(r))
		b.WriteString("\n")
	}
	ctx.SendChain(message.Text(strings.TrimRight(b.String(), "\n")))
}

func formatRefreshLine(r servers.RefreshResult) string {
	prefix := fmt.Sprintf("[%s]", r.Label)
	if r.Err != nil {
		return fmt.Sprintf("%s ❌ 失败: %v", prefix, r.Err)
	}
	if r.Skipped {
		v := defaultStrIfEmpty(r.NewVersion, "?")
		return fmt.Sprintf("%s ✅ 已是最新 (dataVersion=%s)", prefix, v)
	}
	old := defaultStrIfEmpty(r.OldVersion, "(无)")
	now := defaultStrIfEmpty(r.NewVersion, "(未探测)")
	return fmt.Sprintf("%s 🔄 %s → %s (加载 %d 个文件)", prefix, old, now, r.FilesLoaded)
}

func defaultStrIfEmpty(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
