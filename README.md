# ModernFM - 现代化多功能文件管理系统 🚀

ModernFM 是一款专为 Unraid 和私有云设计的现代化文件管理器。基于 Docker 部署，采用 Go 后端 + Vue 3 前端，并结合 PostgreSQL 与 Redis 提供极致的性能。

## ✨ 核心特性
- **🚀 极速响应**: 基于 Go 语言开发，支持百万级文件秒开。
- **🎨 现代 UI**: 深度复刻 Alist 风格，支持**网格/列表**切换。
- **🌓 主题切换**: 完美支持浅色、深色及跟随系统主题。
- **🎬 影音增强**: 支持 **FFmpeg 实时转码**，支持 VLC/Infuse 一键播放。
- **📝 文档编辑**: 在线预览 PDF/Office，支持 YAML/Markdown 在线编辑。
- **🛡️ 企业级架构**: 
  - **PostgreSQL**: 持久化存储文件元数据，搜索快如闪电。
  - **Redis**: 毫秒级目录缓存。
  - **分块上传**: 支持大文件分块上传与断点续传。

---

## 🛠️ 安装说明 (Docker Compose)

### 1. 克隆仓库
```bash
git clone https://github.com/flywindW666/ModernFM.git
cd ModernFM/deploy
```

### 2. 配置环境变量
复制模板文件并编辑：
```bash
cp .env.example .env
```
**关键参数说明：**
- `HOST_DATA_PATH`: 您想在网页上管理的真实文件路径（如 Unraid 的 `/mnt/user`）。
- `WEB_UI_PORT`: 网页访问端口（默认 80）。
- `API_PORT`: 后端接口端口（默认 38866）。

### 3. 一键部署
```bash
docker-compose up -d
```
*提示：如果是首次运行，系统会自动拉取官方 PostgreSQL/Redis 镜像并构建核心组件。*

---

## ⚙️ 进阶配置 (.env)
| 变量名 | 说明 | 默认值 |
| :--- | :--- | :--- |
| WEB_UI_PORT | 前端访问端口 | 80 |
| API_PORT | 后端 API 端口 | 38866 |
| HOST_DATA_PATH | 宿主机数据挂载路径 | /mnt/user |
| POSTGRES_PASSWORD | 数据库密码 | secure_pass_123 |

---

## 📂 项目结构
- `/backend`: Go 后端源码（包含索引、转码、分块上传逻辑）。
- `/frontend`: Vue 3 + Vite 前端源码。
- `/deploy`: Docker Compose 部署配置及环境模板。

---
*Developed by Lucky 🍀 & flywindW666*
