# Docker 部署说明

## 需要持久化的目录

在 Docker 环境运行 Moebot-NEXT 时，只需要持久化一个目录：

### 必须持久化

1. **`/app/data`** - 数据目录
   - 包含配置文件（`data/config.yml`）
   - 包含数据库文件（如 `moebot.db`、`moebot-validation.db`）
   - 包含 master 数据（`data/master`）
   - 包含缓存数据（`data/cache`）- 预缓存卡图缩略图等（无限期缓存，需持久化）
   - 挂载方式：`./data:/app/data`
## 使用 docker-compose

项目已提供 `docker-compose.yml`，直接运行即可：

```bash
# 启动服务（首次启动会自动从 config.example.yml 创建配置）
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## 目录结构

```
Moebot-NEXT-Go/
└── data/              # 持久化数据目录（必须）
    ├── config.yml     # 配置文件
    ├── moebot.db
    ├── moebot-validation.db
    ├── master/
    └── cache/
```

## 注意事项

- 首次启动时，entrypoint 脚本会自动从 `config.example.yml` 复制配置到 `data/config.yml`
- 如需修改配置，直接编辑 `data/config.yml` 文件
- 数据目录 `/app/data` 已在 Dockerfile 中通过 `VOLUME` 声明
- 控制台「备份恢复」页面可以把 `/app/data` 打包上传到 S3 兼容存储（MinIO / R2 / AWS S3 等）
- 备份默认排除 `data/cache`、备份临时目录、`*.tmp` 与 `.restore-backup-*`，避免把可再生成缓存上传；可在控制台或 `backup.exclude_patterns` 调整
- 使用 MinIO 等兼容服务时，通常填写 `endpoint`（如 `minio.example.com:9000`）并保持 `force_path_style: true`
- 恢复远端备份后请重启容器：`docker compose restart moebot`
- 建议定期备份 `data` 目录；备份包包含 `data/config.yml`，请确保备份桶为私有桶并使用最小权限密钥
