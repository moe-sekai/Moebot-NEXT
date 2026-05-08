package commands

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"moebot-next/internal/config"
	"moebot-next/internal/models"
	"moebot-next/internal/plugins/moesekai/servers"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

// bindCommandHint returns the user-facing bind command for a region.
// JP (default) uses "/绑定"; other regions use "/{region}绑定" (e.g. "/cn绑定").
func bindCommandHint(region string) string {
	region = config.NormalizeRegion(region)
	if region == "" || region == config.RegionJP {
		return "/绑定"
	}
	return "/" + region + "绑定"
}

// requireRuntime resolves the runtime for the current command and short-circuits
// with a unified message when the runtime is nil or disabled.
func requireRuntime(deps *Deps, ctx *zero.Ctx, forcedRegion string) (*servers.Runtime, *models.User, bool) {
	runtime, user := runtimeForCommand(deps, ctx, forcedRegion)
	if runtime == nil || !runtime.Enabled {
		ctx.SendChain(message.Text(runtimeUnavailableText(runtime)))
		return runtime, user, false
	}
	return runtime, user, true
}

// requireRuntimeWithStore additionally requires the runtime's masterdata Store
// to be loaded (commands that hit cards/musics/events directly).
func requireRuntimeWithStore(deps *Deps, ctx *zero.Ctx, forcedRegion string) (*servers.Runtime, *models.User, bool) {
	runtime, user, ok := requireRuntime(deps, ctx, forcedRegion)
	if !ok {
		return runtime, user, false
	}
	if runtime.Store == nil {
		ctx.SendChain(message.Text(fmt.Sprintf("%s 主数据正在加载中，请稍后再试", runtime.Label)))
		return runtime, user, false
	}
	return runtime, user, true
}

// requireBoundUser ensures the user has a bound game ID for the runtime's region.
// When forcedRegion is non-empty, the user is re-fetched from the database for
// that specific region so the inferred user (which may belong to a different
// region) is not silently reused.
func requireBoundUser(deps *Deps, ctx *zero.Ctx, runtime *servers.Runtime, forcedRegion string, inferredUser *models.User) (*models.User, bool) {
	user := inferredUser
	if forcedRegion != "" {
		var err error
		user, err = deps.DB.GetUserByPlatformRegion("onebot", userIDFromCtx(ctx), runtime.Region)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.SendChain(message.Text("数据库繁忙，请稍后重试"))
			return nil, false
		}
	}
	if user == nil || strings.TrimSpace(user.GameID) == "" {
		hint := bindCommandHint(runtime.Region)
		ctx.SendChain(message.Text(fmt.Sprintf(
			"你还没有绑定 %s 游戏账号~\n请发送 %s 你的游戏ID 来绑定（例：%s 123456789012345678）",
			runtime.Label, hint, hint,
		)))
		return nil, false
	}
	return user, true
}

// requireSekai checks the runtime exposes a usable Sekai API client.
func requireSekai(ctx *zero.Ctx, runtime *servers.Runtime, feature string) bool {
	if runtime == nil || runtime.Sekai == nil || !runtime.Sekai.Enabled() {
		label := "服务器"
		if runtime != nil {
			label = runtime.Label
		}
		ctx.SendChain(message.Text(fmt.Sprintf("暂未配置 %s Sekai API，无法获取%s", label, feature)))
		return false
	}
	return true
}

// requireSuite checks the runtime exposes a usable Haruki public suite client.
func requireSuite(ctx *zero.Ctx, runtime *servers.Runtime, feature string) bool {
	if runtime == nil || runtime.Suite == nil || !runtime.Suite.Enabled() {
		label := "服务器"
		if runtime != nil {
			label = runtime.Label
		}
		ctx.SendChain(message.Text(fmt.Sprintf("暂未配置 %s 抓包数据接口（Haruki 公开 API），无法查询%s", label, feature)))
		return false
	}
	return true
}

// requireSuiteVisible enforces the user has not hidden their suite data.
func requireSuiteVisible(deps *Deps, ctx *zero.Ctx, runtime *servers.Runtime) (*models.SuiteSetting, bool) {
	setting := suiteSettingOrDefault(deps, userIDFromCtx(ctx), runtime.Region)
	if setting.Hidden {
		ctx.SendChain(message.Text(fmt.Sprintf(
			"你已隐藏 %s 抓包信息，请发送 /%s展示抓包 重新展示",
			runtime.Label, runtime.Region,
		)))
		return setting, false
	}
	return setting, true
}

// requireArgument ensures a non-empty trimmed argument was provided. The hint is
// shown to the user when missing (e.g. "/查卡 1204" or "角色名 [all 任务名]").
func requireArgument(ctx *zero.Ctx, raw string, hint string) (string, bool) {
	arg := strings.TrimSpace(raw)
	if arg == "" {
		ctx.SendChain(message.Text(fmt.Sprintf("请提供参数：%s", hint)))
		return "", false
	}
	return arg, true
}

// fetchSuiteUserData calls Suite.GetUserData and translates failures into a
// user-friendly message via friendlySuiteError. Returns ok=true on success.
func fetchSuiteUserData(ctx *zero.Ctx, runtime *servers.Runtime, gameID string, feature string, fields []string, out any) bool {
	if err := runtime.Suite.GetUserData(gameID, "", fields, out); err != nil {
		ctx.SendChain(message.Text(friendlySuiteError(runtime, feature, err)))
		return false
	}
	return true
}

// friendlySuiteError converts a Suite client error into a friendly Chinese
// message without leaking raw URLs, status lines or stack details.
func friendlySuiteError(runtime *servers.Runtime, feature string, err error) string {
	if err == nil {
		return ""
	}
	label := "服务器"
	if runtime != nil {
		label = runtime.Label
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "suite api is disabled"):
		return fmt.Sprintf("暂未配置 %s 抓包数据接口（Haruki 公开 API），无法查询%s", label, feature)
	case strings.Contains(msg, "uid is empty"):
		return fmt.Sprintf("游戏 ID 为空，请先绑定 %s 账号后再试", label)
	case strings.Contains(msg, "returned 404"):
		return fmt.Sprintf("尚未在 Haruki Suite 上传 %s 抓包数据，请先在抓包工具完成上传后再试", label)
	case strings.Contains(msg, "returned 401"), strings.Contains(msg, "returned 403"):
		return fmt.Sprintf("%s 抓包数据接口鉴权失败，请联系管理员检查 Haruki API 配置", label)
	case strings.Contains(msg, "returned 429"):
		return fmt.Sprintf("%s 抓包数据接口请求过于频繁，请稍后重试", label)
	case isStatus5xx(msg):
		return fmt.Sprintf("%s 抓包数据接口服务异常，请稍后重试", label)
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline exceeded"):
		return fmt.Sprintf("连接 %s 抓包数据接口超时，请稍后重试", label)
	case isNetworkError(err), strings.Contains(msg, "request failed"), strings.Contains(msg, "no such host"), strings.Contains(msg, "connection refused"):
		return fmt.Sprintf("无法连接 %s 抓包数据接口，请稍后重试", label)
	case strings.Contains(msg, "decode suite response"), strings.Contains(msg, "read suite response"):
		return fmt.Sprintf("解析 %s 抓包数据失败，请稍后重试", label)
	case strings.Contains(msg, "parse suite url"):
		return fmt.Sprintf("%s 抓包数据接口配置异常，请联系管理员", label)
	default:
		return fmt.Sprintf("获取%s抓包数据失败，请稍后重试", feature)
	}
}

// friendlySekaiError converts a Sekai client error into a friendly Chinese message.
func friendlySekaiError(runtime *servers.Runtime, feature string, err error) string {
	if err == nil {
		return ""
	}
	label := "服务器"
	if runtime != nil {
		label = runtime.Label
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "sekai api is disabled"):
		return fmt.Sprintf("暂未配置 %s Sekai API，无法获取%s", label, feature)
	case strings.Contains(msg, "user id is empty"):
		return fmt.Sprintf("游戏 ID 为空，请先绑定 %s 账号后再试", label)
	case strings.Contains(msg, "returned 404"):
		return fmt.Sprintf("未在 %s 找到该账号的%s（可能未公开或 ID 有误）", label, feature)
	case strings.Contains(msg, "returned 401"), strings.Contains(msg, "returned 403"):
		return fmt.Sprintf("%s Sekai API 鉴权失败，请联系管理员", label)
	case strings.Contains(msg, "returned 429"):
		return fmt.Sprintf("%s Sekai API 请求过于频繁，请稍后重试", label)
	case isStatus5xx(msg):
		return fmt.Sprintf("%s Sekai API 服务异常，请稍后重试", label)
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline exceeded"):
		return fmt.Sprintf("连接 %s Sekai API 超时，请稍后重试", label)
	case isNetworkError(err), strings.Contains(msg, "request failed"), strings.Contains(msg, "no such host"), strings.Contains(msg, "connection refused"):
		return fmt.Sprintf("无法连接 %s Sekai API，请稍后重试", label)
	default:
		return fmt.Sprintf("获取%s失败，请稍后重试", feature)
	}
}

func isStatus5xx(msg string) bool {
	for _, code := range []string{"returned 500", "returned 501", "returned 502", "returned 503", "returned 504"} {
		if strings.Contains(msg, code) {
			return true
		}
	}
	return false
}

func isNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	return false
}
