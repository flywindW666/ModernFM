# ModernFM - 现代化多功能文件管理系统 🚀

ModernFM 是一款专为 Unraid 和私有云设计的、基于 Docker 部署的现代化文件管理器。从底层重新构建，提供极致的性能与丝滑的 UI 体验。

## ✨ 核心特性
- **极速后端**: 基于 Go 语言开发，支持百万级文件瞬间加载。
- **现代化 UI**: 基于 Vue 3 + Tailwind CSS，支持**网格/列表**切换。
- **暗黑模式**: 支持浅色、深色及跟随系统主题。
- **全格式支持**: 在线预览图片、视频、PDF、Office文档，支持 YAML/Markdown 在线编辑。
- **企业级增强**: 
  - 集成 **PostgreSQL** 元数据管理。
  - 集成 **Redis** 高速缓存。
  - 支持 **FFmpeg** 实时视频转码。
  - 支持 **VLC/Infuse** 等第三方播放器一键联动。
- **文件分享**: 支持带提取码和过期时间的分享链接。

## 🚀 快速开始

### 1. 克隆项目
```bash
git clone https://github.com/flywindW666/ModernFM.git
cd ModernFM/deploy
```

### 2. 配置环境
复制 `.env.example` 为 `.env` 并根据需要修改端口及路径：
```bash
cp .env.example .env
# 使用 nano 或 vim 修改 .env 文件
```

### 3. 一键启动
```bash
docker-compose up -d --build
```

## ⚙️ 部署参数 (.env)
您可以轻松自定义以下部署参数：
- `WEB_UI_PORT`: 网页访问端口（默认 80）。
- `API_PORT`: 后端接口端口（默认 38866）。
- `HOST_DATA_PATH`: 您想管理的真实文件路径。

## 🤝 贡献与反馈
欢迎提交 Issue 或 Pull Request！

---
*Developed by Lucky 🍀 & ryan*
