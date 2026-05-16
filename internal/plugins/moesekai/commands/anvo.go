package commands

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"moebot-next/internal/plugins/moesekai/assets"
	"moebot-next/internal/bot"
	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/masterdata"
	"moebot-next/internal/renderer"
	"moebot-next/internal/plugins/moesekai/suite"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"moebot-next/internal/plugins/moesekai/renderpayloads"
)

type anvoProfile struct {
	suite.BaseProfile
	UserGamedata    suite.UserGamedata    `json:"userGamedata"`
	UserDecks       []suite.UserDeck      `json:"userDecks"`
	UserCards       []suite.UserCard      `json:"userCards"`
	UserMusicVocals []userMusicVocal      `json:"userMusicVocals"`
	UserMusics      []userMusicWithVocals `json:"userMusics"`
}

type userMusicVocal struct {
	MusicVocalID int `json:"musicVocalId"`
}

type userMusicWithVocals struct {
	UserMusicVocals []userMusicVocal `json:"userMusicVocals"`
}

type anvoEntry struct {
	MusicVocalID int    `json:"musicVocalId"`
	MusicID      int    `json:"musicId"`
	Title        string `json:"title"`
	PublishedAt  int64  `json:"publishedAt"`
	CharacterIDs []int  `json:"characterIds"`
	CoverURL     string `json:"coverUrl,omitempty"`
	Owned        bool   `json:"owned"`
}

type anvoListPayload struct {
	Title       string      `json:"title"`
	Subtitle    string      `json:"subtitle,omitempty"`
	Profile     any         `json:"profile"`
	CharacterID int         `json:"characterId"`
	Entries     []anvoEntry `json:"entries"`
	OwnedCount  int         `json:"ownedCount"`
	TotalCount  int         `json:"totalCount"`
	AssetSource string      `json:"assetSource,omitempty"`
}

func anvoFields() []string {
	return suite.Fields(suite.FieldUserMusicVocals, suite.FieldUserMusics)
}

func RegisterAnvo(deps *Deps) {
	for _, cmd := range parserCommands(deps, "ANVO持有") {
		commandName := cmd.Name
		forcedRegion := cmd.Region
		Engine.OnCommand(commandName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
			start := time.Now()
			runtime, inferredUser, ok := requireRuntimeWithStore(deps, ctx, forcedRegion)
			if !ok {
				return
			}
			user, ok := requireBoundUser(deps, ctx, runtime, forcedRegion, inferredUser)
			if !ok {
				return
			}
			if !requireSuite(ctx, runtime, "ANVO持有") {
				return
			}
			if _, ok := requireSuiteVisible(deps, ctx, runtime); !ok {
				return
			}
			cid, err := parseAnvoArgs(commandArgs(ctx))
			if err != nil {
				ctx.SendChain(message.Text(err.Error()))
				return
			}
			var profile anvoProfile
			if !fetchSuiteUserData(ctx, runtime, user.GameID, "ANVO持有", anvoFields(), &profile) {
				return
			}
			entries := buildAnvoEntries(runtime.Store, runtime.Assets, cid, ownedMusicVocalIDs(profile), time.Now())
			if len(entries) == 0 {
				ctx.SendChain(message.Text("该角色暂无可查询的 Another Vocal"))
				return
			}
			payload := buildAnvoPayload(runtime.Region, cid, profile, entries, runtime.Assets)
			sendAnvoOrText(ctx, deps, payload, formatAnvoText(runtime.Region, cid, profile, entries))
			bot.RecordCommandRegion(deps.DB, "ANVO持有", runtime.Region, ctx, start)
		})
	}
}

func parseAnvoArgs(raw string) (int, error) {
	arg := strings.TrimSpace(raw)
	if arg == "" {
		return 0, errors.New("使用方式: /anvo 角色名")
	}
	fields := strings.Fields(arg)
	if len(fields) != 1 {
		return 0, fmt.Errorf("参数无法解析: %s\n使用方式: /anvo 角色名", strings.Join(fields[1:], " "))
	}
	cid := characterIDByAlias(fields[0])
	if cid <= 0 {
		return 0, fmt.Errorf("角色名无效: %s", fields[0])
	}
	return cid, nil
}

func ownedMusicVocalIDs(profile anvoProfile) map[int]bool {
	owned := map[int]bool{}
	for _, vocal := range profile.UserMusicVocals {
		if vocal.MusicVocalID > 0 {
			owned[vocal.MusicVocalID] = true
		}
	}
	for _, music := range profile.UserMusics {
		for _, vocal := range music.UserMusicVocals {
			if vocal.MusicVocalID > 0 {
				owned[vocal.MusicVocalID] = true
			}
		}
	}
	return owned
}

func buildAnvoEntries(store *masterdata.Store, resolver *assets.Resolver, cid int, owned map[int]bool, now time.Time) []anvoEntry {
	if store == nil || cid <= 0 {
		return nil
	}
	assetResolver := resolver
	musicByID := map[int]masterdata.MusicInfo{}
	for _, music := range store.AllMusics() {
		if music.PublishedAt > 0 && music.PublishedAt > now.UnixMilli() {
			continue
		}
		musicByID[music.ID] = music
	}
	entries := make([]anvoEntry, 0)
	for _, vocal := range store.AllMusicVocals() {
		if vocal.MusicVocalType != "" && vocal.MusicVocalType != "another_vocal" {
			continue
		}
		music, ok := musicByID[vocal.MusicID]
		if !ok {
			continue
		}
		characterIDs := vocalCharacterIDs(vocal)
		if !containsInt(characterIDs, cid) {
			continue
		}
		entry := anvoEntry{MusicVocalID: vocal.ID, MusicID: vocal.MusicID, Title: music.Title, PublishedAt: music.PublishedAt, CharacterIDs: characterIDs, Owned: owned[vocal.ID]}
		if assetResolver != nil && music.AssetbundleName != "" {
			entry.CoverURL = assetResolver.GetMusicJacketURL(music.AssetbundleName)
		}
		entries = append(entries, entry)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].PublishedAt != entries[j].PublishedAt {
			return entries[i].PublishedAt < entries[j].PublishedAt
		}
		if entries[i].MusicID != entries[j].MusicID {
			return entries[i].MusicID < entries[j].MusicID
		}
		return entries[i].MusicVocalID < entries[j].MusicVocalID
	})
	return entries
}

func vocalCharacterIDs(vocal masterdata.MusicVocal) []int {
	chars := append([]masterdata.MusicVocalCharacter(nil), vocal.Characters...)
	sort.SliceStable(chars, func(i, j int) bool { return chars[i].Seq < chars[j].Seq })
	out := make([]int, 0, len(chars))
	seen := map[int]struct{}{}
	for _, ch := range chars {
		if ch.CharacterID <= 0 || ch.CharacterType != "game_character" {
			continue
		}
		if _, ok := seen[ch.CharacterID]; ok {
			continue
		}
		seen[ch.CharacterID] = struct{}{}
		out = append(out, ch.CharacterID)
	}
	return out
}

func buildAnvoPayload(region string, cid int, profile anvoProfile, entries []anvoEntry, resolver *assets.Resolver) anvoListPayload {
	owned := 0
	for _, entry := range entries {
		if entry.Owned {
			owned++
		}
	}
	assetSource := ""
	if resolver != nil {
		assetSource = resolver.RendererAssetSource()
	}
	return anvoListPayload{
		Title:       fmt.Sprintf("%s Another Vocal 持有情况", characterDisplayName(cid)),
		Subtitle:    fmt.Sprintf("已持有 %d/%d 首", owned, len(entries)),
		Profile:     renderpayloads.BuildSuiteProfilePayload(region, "anvo", profile.BaseProfile, profile.UserGamedata),
		CharacterID: cid,
		Entries:     entries,
		OwnedCount:  owned,
		TotalCount:  len(entries),
		AssetSource: assetSource,
	}
}

func sendAnvoOrText(ctx *zero.Ctx, deps *Deps, payload anvoListPayload, fallback string) {
	if deps.Renderer != nil {
		if png, err := deps.Renderer.Render(renderer.RenderRequest{Template: "anvo_list", Data: payload}); err == nil {
			ctx.SendChain(message.ImageBytes(png))
			return
		}
	}
	ctx.SendChain(message.Text(fallback))
}

func formatAnvoText(region string, cid int, profile anvoProfile, entries []anvoEntry) string {
	owned := 0
	for _, entry := range entries {
		if entry.Owned {
			owned++
		}
	}
	name := profile.UserGamedata.Name
	if name == "" {
		name = "未知玩家"
	}
	lines := []string{
		fmt.Sprintf("%s %s Another Vocal", strings.ToUpper(config.NormalizeRegion(region)), characterDisplayName(cid)),
		fmt.Sprintf("玩家: %s", name),
		fmt.Sprintf("已持有: %d/%d", owned, len(entries)),
		fmt.Sprintf("更新时间: %s", suiteUpdateText(profile.UploadTime)),
	}
	for _, entry := range entries {
		status := "未持有"
		if entry.Owned {
			status = "已持有"
		}
		lines = append(lines, fmt.Sprintf("#%d %s · %s", entry.MusicID, entry.Title, status))
	}
	return strings.Join(lines, "\n")
}

func characterIDByAlias(raw string) int {
	key := strings.ToLower(strings.TrimSpace(raw))
	aliases := map[string]int{
		"一歌": 1, "ick": 1, "ichika": 1,
		"咲希": 2, "saki": 2,
		"穗波": 3, "hnm": 3, "honami": 3,
		"志步": 4, "shiho": 4,
		"实乃理": 5, "mnr": 5, "minori": 5,
		"遥": 6, "hrk": 6, "haruka": 6,
		"爱莉": 7, "airi": 7,
		"雫": 8, "szk": 8, "shizuku": 8,
		"心羽": 9, "khn": 9, "kohane": 9,
		"杏": 10, "an": 10,
		"彰人": 11, "akt": 11, "akito": 11,
		"冬弥": 12, "toya": 12,
		"司": 13, "tks": 13, "tsukasa": 13,
		"笑梦": 14, "emu": 14,
		"宁宁": 15, "nene": 15,
		"类": 16, "rui": 16,
		"奏": 17, "knd": 17, "kanade": 17,
		"真冬": 18, "mfy": 18, "mafuyu": 18,
		"绘名": 19, "ena": 19,
		"瑞希": 20, "mzk": 20, "mizuki": 20,
		"初音未来": 21, "miku": 21,
		"镜音铃": 22, "rin": 22,
		"镜音连": 23, "len": 23,
		"巡音流歌": 24, "luka": 24,
		"meiko": 25,
		"kaito": 26,
	}
	return aliases[key]
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
