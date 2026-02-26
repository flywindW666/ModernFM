# ModernFM - 现代化多功能文件管理系统 🚀

ModernFM 是一款专为 Unraid 和私有云设计的现代化文件管理器。采用 **All-in-One** 架构，单镜像集成 Go 后端与 Vue 3 前端，部署极简。

## ✨ 核心特性
- **📦 All-in-One**: 单个 Docker 镜像即可提供完整的 Web UI 和 API 服务。
- **🚀 极速响应**: 基于 Go 语言开发，配合 PostgreSQL 索引，支持百万级文件秒开。
- **🎨 现代 UI**: 深度复刻 Alist 风格，支持**网格/列表**切换。
- **🌓 响应式设计**: 完美支持浅色/深色模式及移动端适配。
- **🎬 影音增强**: 内置 **FFmpeg 实时转码**，支持跳转 VLC/Infuse/PotPlayer 播放。
- **📝 全能管理**: 支持大文件**分块上传**、ZIP/TAR 在线压缩与解压。
- **🛡️ 稳健后端**: 
  - **PostgreSQL**: 持久化存储元数据，搜索快如闪电。
  - **Redis**: 毫秒级目录缓存。

---

## 🛠️ 快速开始 (Docker Compose)

我们已经将所有配置集成到了 Compose 文件中，无需配置 `.env`，真正实现开箱即用。

### 1. 克隆并进入目录
```bash
git clone https://github.com/flywindW666/ModernFM.git
cd ModernFM/deploy
```

### 2. (可选) 修改数据挂载路径
编辑 `docker-compose.yml`，将 `/mnt/user` 修改为您真实的媒体/数据存放路径：
```yaml
volumes:
  - /mnt/user:/data  # 将左侧改为您的路径
```

### 3. 一键启动
```bash
docker-compose up -d
```

---

## 🔗 访问信息
部署完成后，直接访问后端端口即可进入系统：

- **Web 界面 & API**: `http://<服务器IP>:38866`

---

## 📂 项目结构
- `/backend`: Go 后端源码（索引、转码、分块上传）。
- `/frontend`: Vue 3 + Vite 前端源码。
- `/deploy`: Docker Compose 一键部署脚本及配置文件。
- `Dockerfile.all-in-one`: 自动化构建前后端集成镜像的定义文件。

---
*Developed by Lucky 🍀 & flywindW666*
