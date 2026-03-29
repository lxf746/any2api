# Any2API

Any2API 是一个多 provider 的 `to-api` 网关项目，用统一入口桥接 `cursor`、`kiro`、`grok`、`orchids`、`web`、`chatgpt`、`blink`，并同时提供：

- OpenAI 兼容接口：`/v1/chat/completions`
- Anthropic 兼容接口：`/v1/messages`
- 多媒体接口：`/v1/images/generations`、`/v1/audio/speech`、`/v1/ocr`
- 管理接口：`/api/admin/*` / `/admin/api/*`（不同语言版本略有差异）

## 适合做什么

- 把多个上游模型/平台统一暴露成一套 API
- 在 Go / Python / Rust 之间对比实现策略
- 继续扩展账号池、鉴权、流式转发和管理后台

## 当前内置 provider

- `cursor`
- `kiro`
- `grok`
- `orchids`
- `web`
- `chatgpt`
- `blink`

## 仓库结构

- `any2api/go`：Any2API Go 后端，当前最完整，带内置 Web Admin
- `any2api/python`：Any2API Python 后端，提供兼容 API 与管理接口
- `any2api/rust`：Any2API Rust 后端，提供兼容 API 与管理接口
- `any2api/desktop`：Any2API Desktop，统一 Tauri 管理端

## 当前状态

### Go

- 最完整的后端实现
- 提供 `/admin` Web 管理台
- 对 provider capability / execution interface 的抽象最完整
- Cursor 已有真实上游链路，支持流式与非流式
- 已补齐 Blink 的真实上游链路、Firebase 刷新、项目创建和 OpenAI / Anthropic 兼容转发
- 已补齐 Z.ai Image / TTS / OCR 的公开端点与后台配置

### Python

- 提供健康检查、模型列表、OpenAI/Anthropic 兼容接口
- 已内置 `cursor` / `kiro` / `grok` / `orchids` / `web` / `chatgpt`
- 提供管理登录、会话与配置接口
- 已补齐 `web` / `chatgpt` 的 provider 配置与管理路由
- 已补齐 Z.ai Image / TTS / OCR 的公开端点、管理配置持久化与状态展示
- 适合作为轻量实现和协议对照版本

### Rust

- 提供健康检查、模型列表、OpenAI/Anthropic 兼容接口
- 已补齐 `cursor` / `kiro` / `grok` / `orchids` / `web` / `chatgpt` 的真实 provider 主链路
- 已补齐 `web` / `chatgpt` 的 provider 配置与管理路由
- 已补齐 Z.ai Image / TTS / OCR 的公开端点、管理配置持久化与状态展示
- 当前 `stream=true` 仍以服务端包装 SSE 为主，不是 Go 那种 provider 原生流式透传

## 快速启动

### Go

1. `cd any2api/go`
2. `go run ./cmd/server`
3. 默认地址：`http://127.0.0.1:8099`
4. Admin 页面：`http://127.0.0.1:8099/admin`

### Python

1. `cd any2api/python`
2. `python3 server.py`
3. 默认地址：`http://127.0.0.1:8100`
4. 可通过 `NEWPLATFORM2API_HOST` / `HOST` 和 `NEWPLATFORM2API_PORT` / `PORT` 覆盖监听地址

### Rust

1. `cd any2api/rust`
2. `cargo run`
3. 默认地址：`http://127.0.0.1:8101`
4. 可通过 `NEWPLATFORM2API_BIND_ADDR` / `BIND_ADDR` 或 `NEWPLATFORM2API_HOST` / `HOST` + `NEWPLATFORM2API_PORT` / `PORT` 覆盖监听地址

## Docker 部署

- 已提供：
  - `any2api/go/Dockerfile`
  - `any2api/python/Dockerfile`
  - `any2api/rust/Dockerfile`
  - `any2api/docker-compose.yml`
- 启动：
  1. `cd any2api`
  2. `docker compose up --build -d`
- 默认暴露端口：
  - Go：`8099`
  - Python：`8100`
  - Rust：`8101`
- 数据持久化目录：
  - `./go/data`
  - `./python/data`
  - `./rust/data`
- 默认访问地址：
  - Go：`http://127.0.0.1:8099`
  - Python：`http://127.0.0.1:8100`
  - Rust：`http://127.0.0.1:8101`

## 通用接口

- `GET /health`
- `GET /v1/models`
- `POST /v1/chat/completions`
- `POST /v1/messages`
- `POST /v1/images/generations`
- `POST /v1/audio/speech`
- `POST /v1/ocr`

## Z.ai 多媒体能力

当前 Go / Python / Rust 都已对齐这三类能力，并可通过管理接口或环境变量配置凭证：

- **图片生成**：`POST /v1/images/generations`
  - 需要 `prompt`
  - 当前只支持 `n=1`
  - 当前只支持 `response_format=url`
  - `size` 映射：
    - `1024x1024 -> ratio=1:1, resolution=1K`
    - `1024x1792 -> ratio=9:16, resolution=2K`
    - `1792x1024 -> ratio=16:9, resolution=2K`
- **语音合成**：`POST /v1/audio/speech`
  - 支持 `input` 或 `text`
  - 当前只支持 `response_format=wav`
  - 默认语音：`voice_id=system_003`、`voice_name=通用男声`
- **OCR 上传**：`POST /v1/ocr`
  - `multipart/form-data`
  - 需要 `file` 字段
  - 返回统一 OCR 结果对象：`text` / `markdown` / `json` / `layout` / `file`

示例：

- 图片生成：`curl -X POST http://127.0.0.1:8099/v1/images/generations -H 'Content-Type: application/json' -d '{"prompt":"draw a cat","size":"1024x1024"}'`
- 语音合成：`curl -X POST http://127.0.0.1:8099/v1/audio/speech -H 'Content-Type: application/json' -d '{"input":"hello world"}' --output speech.wav`
- OCR：`curl -X POST http://127.0.0.1:8099/v1/ocr -F 'file=@example.png'`

## 管理能力

- 默认管理密码：`changeme`
- 运行时配置会持久化到各语言目录下的 `data/admin.json`
- Go 版带内置管理页面
- Desktop 已支持加载/保存 Z.ai Image / TTS / OCR 配置
- Python / Rust 当前以管理 API 为主
- Python / Rust 已支持：`GET/PUT /admin/api/providers/web/config`
- Python / Rust 已支持：`GET/PUT /admin/api/providers/chatgpt/config`
- Go / Python / Rust 已支持：
  - `GET/PUT /admin/api/providers/zai/image/config`
  - `GET/PUT /admin/api/providers/zai/tts/config`
  - `GET/PUT /admin/api/providers/zai/ocr/config`
- 运行时 provider 配置 key：
  - `webConfig`
  - `chatgptConfig`
  - `zaiImageConfig`
  - `zaiTTSConfig`
  - `zaiOCRConfig`

常见环境变量（为了兼容现有部署，前缀暂时保留为旧名）：

- `NEWPLATFORM2API_API_KEY`
- `NEWPLATFORM2API_ADMIN_PASSWORD`
- `NEWPLATFORM2API_DEFAULT_PROVIDER`
- `NEWPLATFORM2API_DATA_DIR`
- `NEWPLATFORM2API_HOST`
- `NEWPLATFORM2API_PORT`
- `NEWPLATFORM2API_BIND_ADDR`

Z.ai 多媒体相关环境变量：

- Image
  - `NEWPLATFORM2API_ZAI_IMAGE_SESSION_TOKEN`
  - `ZAI_IMAGE_SESSION_TOKEN`
  - `NEWPLATFORM2API_ZAI_IMAGE_API_URL`
- TTS
  - `NEWPLATFORM2API_ZAI_TTS_TOKEN`
  - `ZAI_TTS_TOKEN`
  - `NEWPLATFORM2API_ZAI_TTS_USER_ID`
  - `ZAI_TTS_USER_ID`
  - `NEWPLATFORM2API_ZAI_TTS_API_URL`
- OCR
  - `NEWPLATFORM2API_ZAI_OCR_TOKEN`
  - `ZAI_OCR_TOKEN`
  - `NEWPLATFORM2API_ZAI_OCR_API_URL`

## Desktop

`any2api/desktop` 是 Any2API 的统一 Tauri 管理端，目标是复用同一套管理协议，而不是为 Go / Python / Rust 分别维护三套桌面 UI。

更多说明见：`any2api/desktop/README.md`

## 改名说明

项目品牌名已调整为 **Any2API**，但以下兼容项目前仍保留：

- 仓库目录现已调整为 `any2api`
- 现有环境变量前缀仍为 `NEWPLATFORM2API_`
- 现有部分 cookie / 本地存储 key 维持旧值，避免直接破坏现有登录态与本地配置

## 下一步建议

如果你要继续把它往生产可用推进，建议优先处理：

1. 更完整的 provider 原生流式转发
2. 多账号池与故障切换
3. 更细的错误映射与限流
4. 更完整的管理后台能力
5. 真实上游鉴权刷新与可观测性
