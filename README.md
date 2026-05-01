# Moebot NEXT

新一代 PJSK (Project SEKAI) 查询机器人。

## 特性

- 🎴 卡牌 / 曲目 / 活动 / 卡池查询
- 🖼️ 精美图片卡片渲染 (Satori, 无需 Chromium)
- 🔗 OneBot 协议，兼容 NapCat / Lagrange / LLOneBot
- 🌐 内置管理面板 (Koishi Console)
- 🗄️ 轻量嵌入式数据库 (libSQL/SQLite)
- 🔌 可选接入 SEKAI API 获取实时数据
- 📦 一键部署

## 快速开始

### Docker 部署 (推荐)

```bash
# 1. 克隆仓库
git clone https://github.com/xxx/moebot-next.git
cd moebot-next

# 2. 创建配置文件
cp koishi.example.yml koishi.yml
# 编辑 koishi.yml，设置你的 QQ 机器人 selfId

# 3. 启动
docker compose up -d

# 4. 访问管理面板
# http://localhost:5140
```

### Windows 部署

```batch
:: 1. 安装 Node.js 20+ (https://nodejs.org/)
:: 2. 克隆仓库
git clone https://github.com/xxx/moebot-next.git
cd moebot-next

:: 3. 运行启动脚本
scripts\start.bat
```

### macOS / Linux 部署

```bash
# 1. 安装 Node.js 20+
# 2. 克隆仓库
git clone https://github.com/xxx/moebot-next.git
cd moebot-next

# 3. 运行启动脚本
chmod +x scripts/start.sh
./scripts/start.sh
```

### 一键安装 (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/xxx/moebot-next/main/scripts/install.sh | bash
```

## 配置

编辑 `koishi.yml`，主要配置项：

| 配置 | 说明 | 默认值 |
|------|------|--------|
| `adapter-onebot.selfId` | QQ 机器人账号 | (必填) |
| `moebot-core.masterDataUrl` | Masterdata 数据源 | `https://sk.exmeaning.com/master` |
| `moebot-core.sekaiApi` | SEKAI API 配置 (可选) | 禁用 |

更多配置也可以通过管理面板 WebUI 修改。

## SEKAI API (可选)

接入 SEKAI API 可解锁更多功能：
- 实时排行查询
- 玩家详细数据查询
- Best 30 成绩展示

**不接入也不影响基础查询功能。**

可在管理面板的「SEKAI API」页面配置端点地址和请求头。

## 指令列表

| 指令 | 说明 | 需要 SEKAI API |
|------|------|:--------------:|
| `/查卡 <关键词>` | 搜索卡牌 | ❌ |
| `/查曲 <关键词>` | 搜索曲目 | ❌ |
| `/查活动 [关键词]` | 搜索活动 | ❌ |
| `/查卡池 [关键词]` | 搜索卡池 | ❌ |
| `/表情 <编号>` | 表情贴纸 | ❌ |
| `/绑定 <游戏ID>` | 绑定游戏账号 | ❌ |
| `/个人信息` | 查看个人数据 | ⭕ 可选 |
| `/排行 [排名]` | 实时排行 | ✅ |
| `/b30` | Best 30 | ✅ |
| `/帮助` | 帮助信息 | ❌ |

## 架构

```
moebot-next/
├── packages/
│   ├── shared/     # 共享类型和工具
│   ├── renderer/   # Satori 图片渲染引擎
│   ├── core/       # Koishi 插件 (指令/服务)
│   └── console/    # 管理面板 (Vue)
├── assets/         # 静态资源
├── data/           # 运行时数据 (数据库/缓存)
└── scripts/        # 启动/部署脚本
```

## License

GPL-3.0
