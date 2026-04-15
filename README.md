# Any2API

> ⚠️ **免责声明**：本项目仅供学习和研究使用，不得用于任何商业用途。使用本项目所产生的一切后果由使用者自行承担。

多平台 AI 模型统一 API 网关，将 Cursor、Kiro、Grok、ChatGPT、Blink、Orchids 等平台桥接为 OpenAI / Anthropic 兼容接口。

## 功能特性

- OpenAI 兼容接口：`/v1/chat/completions`、`/v1/models`
- Anthropic 兼容接口：`/v1/messages`
- 多媒体接口：图片生成、语音合成、OCR
- 多 provider 账号池，支持轮询和故障切换
- 内置 Web 管理后台（Go 版）
- 三种语言实现：Go（推荐）、Python、Rust

## 内置 Provider

| Provider | 说明 |
|----------|------|
| `cursor` | Cursor AI 编辑器 |
| `kiro` | Kiro (AWS Builder ID)，支持多账号池 |
| `grok` | Grok (x.ai)，支持多 token 池 |
| `chatgpt` | ChatGPT / OpenAI |
| `blink` | Blink.new |
| `orchids` | Orchids |
| `web` | 通用 OpenAI 兼容后端转发 |
| `zai` | Z.ai 图片生成 / 语音合成 / OCR |

## 快速开始

### Go（推荐）

```bash
cd go
go run ./cmd/server
```

访问 `http://localhost:8099`，管理后台 `http://localhost:8099/admin`，默认密码 `changeme`。

### Python

```bash
cd python
pip install -r requirements.txt
python3 server.py
```

访问 `http://localhost:8100`。

### Rust

```bash
cd rust
cargo run
```

访问 `http://localhost:8101`。

### Docker 一键部署

```bash
docker compose up -d
```

| 服务 | 端口 | 地址 |
|------|------|------|
| Go | 8099 | `http://localhost:8099` |
| Python | 8100 | `http://localhost:8100` |
| Rust | 8101 | `http://localhost:8101` |

数据持久化到各语言目录下的 `data/` 目录。

只启动某个版本：

```bash
docker compose up any2api-go -d     # 只启动 Go
docker compose up any2api-python -d # 只启动 Python
docker compose up any2api-rust -d   # 只启动 Rust
```

## API 使用

### Chat Completions（OpenAI 兼容）

```bash
curl http://localhost:8099/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 0000" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'
```

### Messages（Anthropic 兼容）

```bash
curl http://localhost:8099/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: 0000" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "messages": [{"role": "user", "content": "Hello"}],
    "max_tokens": 1024
  }'
```

### 图片生成

```bash
curl http://localhost:8099/v1/images/generations \
  -H "Content-Type: application/json" \
  -d '{"prompt": "a cat", "size": "1024x1024"}'
```

### 语音合成

```bash
curl http://localhost:8099/v1/audio/speech \
  -H "Content-Type: application/json" \
  -d '{"input": "hello world"}' \
  --output speech.wav
```

### OCR

```bash
curl http://localhost:8099/v1/ocr -F 'file=@image.png'
```

## 管理接口

通过 Web 管理后台或 API 管理 provider 配置和账号池：

| 接口 | 说明 |
|------|------|
| `POST /admin/api/login` | 管理员登录 |
| `GET/PUT /admin/api/settings` | 全局设置（API Key、默认 provider） |
| `GET/PUT /admin/api/providers/cursor/config` | Cursor 配置 |
| `GET/POST /admin/api/providers/kiro/accounts/*` | Kiro 账号池 CRUD |
| `GET/POST /admin/api/providers/grok/tokens/*` | Grok token 池 CRUD |
| `GET/PUT /admin/api/providers/chatgpt/config` | ChatGPT 配置 |
| `GET/PUT /admin/api/providers/blink/config` | Blink 配置 |
| `GET/PUT /admin/api/providers/orchids/config` | Orchids 配置 |
| `GET/PUT /admin/api/providers/web/config` | Web 通用后端配置 |
| `GET/PUT /admin/api/providers/zai/image/config` | Z.ai 图片配置 |
| `GET/PUT /admin/api/providers/zai/tts/config` | Z.ai 语音配置 |
| `GET/PUT /admin/api/providers/zai/ocr/config` | Z.ai OCR 配置 |

## 配合 Any Auto Register 使用

配合 [Any Auto Register](https://github.com/lxf746/any-auto-register) 项目，可实现批量注册账号后自动推送到 Any2API，注册即可用：

1. 在 Any Auto Register 的全局配置中填写 Any2API 地址和管理密码
2. 注册账号时，成功后自动调用 Any2API 管理 API 添加账号
3. 无需手动导出导入，注册完直接通过 `/v1/chat/completions` 使用

也可以在 Any Auto Register 中手动导出 Any2API 格式的 `admin.json`，放到 `data/` 目录下。

## 环境变量

通用配置：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `NEWPLATFORM2API_PORT` / `PORT` | 监听端口 | Go: 8099, Python: 8100, Rust: 8101 |
| `NEWPLATFORM2API_API_KEY` / `API_KEY` | 客户端 API Key | `0000` |
| `NEWPLATFORM2API_ADMIN_PASSWORD` / `ADMIN_PASSWORD` | 管理密码 | `changeme` |
| `NEWPLATFORM2API_DEFAULT_PROVIDER` | 默认 provider | `cursor` |
| `NEWPLATFORM2API_DATA_DIR` / `DATA_DIR` | 数据目录 | `data` |

各 provider 的环境变量以 `NEWPLATFORM2API_{PROVIDER}_` 为前缀，详见各平台接入文档。

## 仓库结构

```
any2api/
├── go/                 # Go 后端（推荐，最完整）
├── python/             # Python 后端
├── rust/               # Rust 后端
├── desktop/            # Tauri 统一管理端
├── docs/               # 平台接入文档 + 逆向教程
├── tools/              # 辅助工具（JS 逆向分析器等）
└── docker-compose.yml  # Docker 编排
```

## 文档

详细的平台接入文档和逆向教程见 `docs/` 目录：

- [项目总览](docs/01-项目总览-any2api介绍.md)
- [Kiro 接入](docs/02-平台接入-Kiro.md)
- [Cursor 接入](docs/03-平台接入-Cursor.md)
- [Grok 接入](docs/04-平台接入-Grok.md)
- [Orchids 接入](docs/05-平台接入-Orchids.md)
- [Web 通用接入](docs/06-平台接入-Web通用.md)
- [Z.ai 接入](docs/07-平台接入-Zai.md)
- [Z.ai 签名算法逆向](docs/08-逆向教程-Zai签名算法.md)
- [逆向分类指南](docs/09-逆向分类指南.md)
- [学习路径指南](docs/10-学习路径指南.md)

## License

本项目采用 [AGPL-3.0](LICENSE) 许可证。个人学习和研究可自由使用；商业使用需遵守 AGPL-3.0 条款（衍生作品须开源）。
