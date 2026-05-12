<template>
  <main class="page-stack">
    <PageHeader
      eyebrow="Groups"
      title="群组管理"
      subtitle="查看已交互群组、命令使用统计，按 OneBot 客户端管理启停状态、备注与每群配置 (JSON)。"
    >
      <template #actions>
        <UiButton
          variant="outline"
          size="sm"
          :loading="loading"
          @click="loadGroups"
          >刷新</UiButton
        >
      </template>
    </PageHeader>

    <UiAlert v-if="success" variant="info" title="操作完成">{{
      success
    }}</UiAlert>
    <UiAlert v-if="error" variant="destructive" title="操作失败">{{
      error
    }}</UiAlert>

    <UiCard>
      <div class="card-heading">
        <div>
          <h2>群组列表</h2>
          <p>
            共 {{ total }} 个群；最近 {{ statsDays }} 天命令统计随列表附带。
          </p>
        </div>
        <div class="groups-toolbar">
          <label class="groups-toolbar__field">
            <span>关键词</span>
            <input
              v-model.trim="filterKeyword"
              class="ui-input"
              placeholder="按群号 / 名称 / 客户端过滤"
            />
          </label>
          <label class="groups-toolbar__field">
            <span>平台</span>
            <select v-model="filterPlatform" class="ui-select">
              <option value="">全部</option>
              <option v-for="opt in platformOptions" :key="opt" :value="opt">
                {{ opt }}
              </option>
            </select>
          </label>
          <label class="groups-toolbar__field">
            <span>客户端</span>
            <select
              v-model="selectedClient"
              class="ui-select"
              @change="reloadFirstPage"
            >
              <option value="__all__">全部客户端</option>
              <option
                v-for="client in clientOptions"
                :key="clientOptionKey(client.client_id)"
                :value="client.client_id"
              >
                {{ formatClient(client.client_id) }}（{{ client.count }} 群）
              </option>
            </select>
          </label>
          <label class="groups-toolbar__field">
            <span>统计天数</span>
            <select
              v-model.number="statsDays"
              class="ui-select"
              @change="loadGroups"
            >
              <option :value="1">1 天</option>
              <option :value="7">7 天</option>
              <option :value="30">30 天</option>
              <option :value="90">90 天</option>
            </select>
          </label>
        </div>
      </div>

      <div v-if="loading" class="groups-skeleton">
        <UiSkeleton v-for="item in 4" :key="item" height="86px" />
      </div>
      <div v-else-if="filteredGroups.length === 0" class="empty-state">
        <div class="empty-state__icon"><SvgIcon name="bot" :size="22" /></div>
        <div>
          <strong>暂无群组数据</strong>
          <p>当机器人在群中收到首条命令时会按 OneBot 客户端自动登记。</p>
        </div>
      </div>
      <div v-else class="groups-list">
        <article
          v-for="group in filteredGroups"
          :key="group.id"
          class="group-row"
          :class="{ 'group-row--disabled': !group.enabled }"
        >
          <header class="group-row__head">
            <div class="group-row__title">
              <UiBadge :variant="group.enabled ? 'success' : 'outline'">{{
                group.enabled ? "启用" : "停用"
              }}</UiBadge>
              <strong>{{ group.name || "未命名群" }}</strong>
              <span class="group-row__sub"
                >{{ formatClient(group.client_id) }} · {{ group.platform }} ·
                {{ group.group_id }}</span
              >
            </div>
            <div class="group-row__stats">
              <div>
                <span>命令数（{{ group.stats?.days ?? statsDays }}d）</span
                ><strong>{{ group.stats?.count ?? 0 }}</strong>
              </div>
              <div>
                <span>平均耗时</span
                ><strong>{{ formatAvgMs(group.stats?.avg_ms) }}</strong>
              </div>
              <div>
                <span>最近活跃</span
                ><strong>{{ formatTime(group.stats?.last_used) }}</strong>
              </div>
              <div>
                <span>登记时间</span
                ><strong>{{ formatTime(group.created_at) }}</strong>
              </div>
            </div>
            <div class="group-row__actions">
              <UiButton
                size="sm"
                variant="outline"
                @click="toggleEnabled(group)"
                :loading="busy[group.id] === 'toggle'"
              >
                {{ group.enabled ? "停用" : "启用" }}
              </UiButton>
              <UiButton
                size="sm"
                variant="ghost"
                @click="toggleExpand(group.id)"
              >
                {{ expandedId === group.id ? "收起" : "编辑" }}
              </UiButton>
              <UiButton
                size="sm"
                variant="ghost"
                @click="loadRecent(group)"
                :loading="busy[group.id] === 'recent'"
              >
                最近命令
              </UiButton>
              <UiButton
                size="sm"
                variant="outline"
                @click="confirmDelete(group)"
                :loading="busy[group.id] === 'delete'"
              >
                删除
              </UiButton>
            </div>
          </header>

          <section v-if="expandedId === group.id" class="group-row__editor">
            <label class="settings-field">
              <span>群备注 / 名称</span>
              <input
                v-model.trim="draft.name"
                class="ui-input"
                placeholder="可选；用于控制台展示"
              />
            </label>
            <label class="settings-field settings-field--full">
              <span
                >群配置（JSON）<small style="opacity: 0.6"
                  >— 留空 / {} 表示默认</small
                ></span
              >
              <textarea
                v-model="draft.config"
                class="ui-input ui-textarea"
                rows="6"
                spellcheck="false"
                placeholder='{"feature_disabled": ["b30"]}'
              ></textarea>
            </label>
            <UiAlert
              v-if="draftJSONError"
              variant="warning"
              title="JSON 无效"
              >{{ draftJSONError }}</UiAlert
            >
            <div class="settings-actions-row">
              <UiButton
                size="sm"
                :disabled="!!draftJSONError"
                :loading="busy[group.id] === 'save'"
                @click="saveDraft(group)"
                >保存修改</UiButton
              >
              <UiButton size="sm" variant="ghost" @click="cancelExpand"
                >取消</UiButton
              >
            </div>
          </section>

          <section v-if="recentMap[group.id]" class="group-row__recent">
            <header>
              <strong>最近 {{ recentMap[group.id]!.length }} 条命令</strong>
            </header>
            <div v-if="recentMap[group.id]!.length === 0" class="muted-text">
              暂无命令记录。
            </div>
            <ul v-else class="group-recent-list">
              <li v-for="row in recentMap[group.id]" :key="row.id">
                <code>{{ row.command }}</code>
                <span class="group-recent-list__meta"
                  >{{ formatClient(row.client_id) }} · 用户 {{ row.user_id }} ·
                  {{ row.region || "-" }} · {{ row.response_ms }}ms ·
                  {{ formatTime(row.created_at) }}</span
                >
                <span v-if="row.args" class="group-recent-list__args">{{
                  row.args
                }}</span>
              </li>
            </ul>
          </section>
        </article>
      </div>

      <footer v-if="totalPages > 1" class="groups-pager">
        <UiButton
          size="sm"
          variant="outline"
          :disabled="page <= 1"
          @click="goPage(page - 1)"
          >上一页</UiButton
        >
        <span>第 {{ page }} / {{ totalPages }} 页</span>
        <UiButton
          size="sm"
          variant="outline"
          :disabled="page >= totalPages"
          @click="goPage(page + 1)"
          >下一页</UiButton
        >
      </footer>
    </UiCard>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import {
  deleteGroup,
  getGroupRecentCommands,
  getGroups,
  updateGroup,
} from "../api/client";
import type {
  GroupClientSummary,
  GroupRecentCommand,
  GroupRow,
} from "../api/types";
import PageHeader from "../components/PageHeader.vue";
import SvgIcon from "../components/icons/SvgIcon.vue";
import UiAlert from "../components/ui/UiAlert.vue";
import UiBadge from "../components/ui/UiBadge.vue";
import UiButton from "../components/ui/UiButton.vue";
import UiCard from "../components/ui/UiCard.vue";
import UiSkeleton from "../components/ui/UiSkeleton.vue";

const allClientsValue = "__all__";
const groups = ref<GroupRow[]>([]);
const clientOptions = ref<GroupClientSummary[]>([]);
const total = ref(0);
const page = ref(1);
const limit = ref(20);
const statsDays = ref(7);
const selectedClient = ref(allClientsValue);
const filterKeyword = ref("");
const filterPlatform = ref("");

const loading = ref(false);
const error = ref("");
const success = ref("");

const expandedId = ref<number | null>(null);
const draft = reactive({ name: "", config: "" });
const recentMap = reactive<Record<number, GroupRecentCommand[]>>({});
const busy = reactive<
  Record<number, "toggle" | "save" | "delete" | "recent" | undefined>
>({});

const totalPages = computed(() =>
  total.value === 0 ? 1 : Math.ceil(total.value / limit.value),
);

const platformOptions = computed(() => {
  const set = new Set<string>();
  for (const g of groups.value) if (g.platform) set.add(g.platform);
  return Array.from(set).sort();
});

const filteredGroups = computed(() => {
  const keyword = filterKeyword.value.toLowerCase();
  return groups.value.filter((g) => {
    if (filterPlatform.value && g.platform !== filterPlatform.value)
      return false;
    if (!keyword) return true;
    return (
      g.group_id.toLowerCase().includes(keyword) ||
      (g.client_id || "").toLowerCase().includes(keyword) ||
      (g.name || "").toLowerCase().includes(keyword)
    );
  });
});

const draftJSONError = computed(() => {
  const text = draft.config.trim();
  if (text === "" || text === "{}") return "";
  try {
    JSON.parse(text);
    return "";
  } catch (err) {
    return err instanceof Error ? err.message : "JSON 解析失败";
  }
});

onMounted(loadGroups);

async function loadGroups() {
  loading.value = true;
  error.value = "";
  try {
    const data = await getGroups(
      page.value,
      limit.value,
      statsDays.value,
      apiSelectedClient(),
    );
    groups.value = data.data ?? [];
    clientOptions.value = data.clients ?? [];
    total.value = data.total ?? 0;
  } catch (err) {
    error.value = normalizeError(err, "加载群组列表失败");
  } finally {
    loading.value = false;
  }
}

function reloadFirstPage() {
  page.value = 1;
  expandedId.value = null;
  loadGroups();
}

function apiSelectedClient(): string | undefined {
  return selectedClient.value === allClientsValue
    ? undefined
    : selectedClient.value;
}

function goPage(next: number) {
  if (next < 1 || next > totalPages.value) return;
  page.value = next;
  expandedId.value = null;
  loadGroups();
}

function toggleExpand(id: number) {
  if (expandedId.value === id) {
    cancelExpand();
    return;
  }
  const target = groups.value.find((g) => g.id === id);
  if (!target) return;
  expandedId.value = id;
  draft.name = target.name || "";
  draft.config = formatConfig(target.config);
}

function cancelExpand() {
  expandedId.value = null;
  draft.name = "";
  draft.config = "";
}

async function saveDraft(group: GroupRow) {
  if (draftJSONError.value) return;
  busy[group.id] = "save";
  error.value = "";
  success.value = "";
  try {
    const payload = {
      name: draft.name,
      config: draft.config.trim() || "{}",
    };
    const result = await updateGroup(group.id, payload);
    Object.assign(group, result.data);
    success.value = `已更新 ${group.name || group.group_id}`;
    cancelExpand();
  } catch (err) {
    error.value = normalizeError(err, "保存失败");
  } finally {
    busy[group.id] = undefined;
  }
}

async function toggleEnabled(group: GroupRow) {
  busy[group.id] = "toggle";
  error.value = "";
  success.value = "";
  try {
    const result = await updateGroup(group.id, { enabled: !group.enabled });
    Object.assign(group, result.data);
    success.value = `${group.name || group.group_id} 已${group.enabled ? "启用" : "停用"}`;
  } catch (err) {
    error.value = normalizeError(err, "切换启停失败");
  } finally {
    busy[group.id] = undefined;
  }
}

async function loadRecent(group: GroupRow) {
  if (recentMap[group.id]) {
    delete recentMap[group.id];
    return;
  }
  busy[group.id] = "recent";
  try {
    const result = await getGroupRecentCommands(group.id, 20);
    recentMap[group.id] = result.data ?? [];
  } catch (err) {
    error.value = normalizeError(err, "加载最近命令失败");
  } finally {
    busy[group.id] = undefined;
  }
}

async function confirmDelete(group: GroupRow) {
  if (
    !window.confirm(
      `确认删除群组 ${group.name || group.group_id}（${formatClient(group.client_id)}）的登记记录？\n该群配置与启停状态会被移除（不会影响命令统计历史）。`,
    )
  )
    return;
  busy[group.id] = "delete";
  error.value = "";
  success.value = "";
  try {
    await deleteGroup(group.id);
    groups.value = groups.value.filter((g) => g.id !== group.id);
    total.value = Math.max(0, total.value - 1);
    success.value = `已删除 ${group.name || group.group_id}`;
  } catch (err) {
    error.value = normalizeError(err, "删除失败");
  } finally {
    busy[group.id] = undefined;
  }
}

function formatConfig(raw: string | undefined) {
  if (!raw) return "";
  const trimmed = raw.trim();
  if (!trimmed || trimmed === "{}") return "";
  try {
    return JSON.stringify(JSON.parse(trimmed), null, 2);
  } catch {
    return trimmed;
  }
}

function clientOptionKey(value: string): string {
  return value === "" ? "__unknown__" : value;
}

function formatClient(value: string | null | undefined): string {
  return value ? `OneBot ${value}` : "未知客户端";
}

function formatTime(value: string | null | undefined) {
  if (!value) return "—";
  return new Date(value).toLocaleString();
}

function formatAvgMs(value: number | undefined) {
  if (!value) return "—";
  return `${Math.round(value)} ms`;
}

function normalizeError(err: unknown, fallback: string) {
  if (err instanceof Error) return `${fallback}：${err.message}`;
  return fallback;
}
</script>
