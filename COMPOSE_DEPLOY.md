# ModernFM Docker Compose 部署文档 🚀

本文档将引导您使用 Docker Compose 一键部署 **ModernFM**。

## 📋 部署架构
本套件包含以下 3 个容器：
1.  **ModernFM (All-in-One)**: 核心服务，包含 Go 后端和 Vue 3 前端托管（端口 38866）。
2.  **Database**: PostgreSQL 15 数据库，存储文件元数据。
3.  **Cache**: Redis 7 缓存，加速目录访问。

---

## 🛠️ 部署步骤

### 1. 准备工作
确保您的系统中已安装 Docker 和 Docker Compose。

### 2. 配置与启动
我们已经将所有配置集成到了 `deploy/docker-compose.yml` 中，您只需确认数据路径：

1.  进入部署目录：
    ```bash
    cd deploy
    ```
2.  （可选）根据需要修改 `docker-compose.yml` 中的数据路径：
    ```yaml
    volumes:
      - /mnt/user:/data  # 将左侧改为您的真实数据路径
    ```
3.  一键启动：
    ```bash
    docker-compose up -d
    ```

---

## 🔗 访问信息
部署完成后，直接访问后端端口即可看到界面：

- **Web UI & API**: `http://<服务器IP>:38866`

---
*Developed & Packaged by Lucky 🍀*
