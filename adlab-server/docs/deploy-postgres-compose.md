# AdLab 单机云服务器部署方案

这套方案面向一台 Linux 云服务器，使用 Docker Compose 启动：

- PostgreSQL
- adlab-server backend
- admin frontend
- Nginx 反向代理

## 1. 服务器要求

- Ubuntu 22.04 或 24.04
- 建议 2C4G 起步
- 开放端口：`22`、`80`、`443`
- 安装 Docker 与 Docker Compose

## 2. 上传项目

将整个 `adlab-server` 目录上传到服务器，例如：

```bash
scp -r ./adlab-server user@your-server:/srv/adlab-server
```

## 3. 准备环境变量

复制一份环境变量模板：

```bash
cd /srv/adlab-server
cp .env.example .env
```

如果你准备部署到阿里云或腾讯云，也可以直接使用对应模板：

```bash
cp .env.aliyun.example .env
```

或：

```bash
cp .env.tencent.example .env
```

至少修改这些值：

- `POSTGRES_PASSWORD`
- `ADLAB_ADMIN_TOKEN`

如果服务器无法直接访问 Docker Hub，可以把：

- `POSTGRES_IMAGE`
- `NGINX_IMAGE`
- `GO_BASE_IMAGE`
- `NODE_BASE_IMAGE`
- `BACKEND_RUNTIME_IMAGE`
- `ADMIN_NGINX_IMAGE`

改成你可访问的镜像源地址。阿里云和腾讯云的示例模板已经分别提供：

- [.env.aliyun.example](/Users/redamancy/Dev/AeroBid/adlab-server/.env.aliyun.example)
- [.env.tencent.example](/Users/redamancy/Dev/AeroBid/adlab-server/.env.tencent.example)

例如你有自己的私有镜像仓库时，可以这样写：

```env
POSTGRES_IMAGE=registry.example.com/library/postgres:16-alpine
NGINX_IMAGE=registry.example.com/library/nginx:1.27-alpine
GO_BASE_IMAGE=registry.example.com/library/golang:1.23-alpine
NODE_BASE_IMAGE=registry.example.com/library/node:20-alpine
BACKEND_RUNTIME_IMAGE=registry.example.com/library/alpine:3.20
ADMIN_NGINX_IMAGE=registry.example.com/library/nginx:1.27-alpine
```

## 4. 启动服务

```bash
docker-compose up -d --build
```

如果你使用的是阿里云或腾讯云自建镜像仓库，建议先把这些基础镜像同步到你自己的仓库：

- `postgres:16-alpine`
- `nginx:1.27-alpine`
- `golang:1.23-alpine`
- `node:20-alpine`
- `alpine:3.20`

这样构建和部署都会更稳定。

启动完成后验证：

```bash
docker-compose ps
curl http://127.0.0.1:8080/health
curl http://127.0.0.1/
```

## 5. 初始化数据

首次部署后，执行一次种子脚本：

```bash
docker-compose run --rm --profile tools seed
```

这会使用单独的一次性工具容器把演示数据写入 PostgreSQL。

## 6. 域名与 HTTPS

当前 `nginx/default.conf` 已支持：

- `/` -> admin frontend
- `/api` -> backend
- `/admin` -> backend
- `/lab` -> backend

正式环境建议：

1. 把域名解析到服务器公网 IP
2. 用 Certbot 或现成网关接入 HTTPS
3. 将 443 证书配置挂进 Nginx

## 7. 常用运维命令

查看状态：

```bash
docker-compose ps
```

查看日志：

```bash
docker-compose logs -f backend
docker-compose logs -f admin
docker-compose logs -f nginx
docker-compose logs -f postgres
```

重启服务：

```bash
docker-compose restart
```

更新部署：

```bash
git pull
docker-compose up -d --build
```

首次准备镜像源之后，也可以先单独验证镜像是否可拉取：

```bash
docker pull "$POSTGRES_IMAGE"
docker pull "$NGINX_IMAGE"
```

如果这两条能成功，通常整套 Compose 就不会再卡在基础镜像层。

## 8. 数据持久化与备份

数据库数据保存在 Docker volume `postgres-data`。

备份建议：

```bash
docker-compose exec postgres pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > backup.sql
```

恢复建议：

```bash
cat backup.sql | docker-compose exec -T postgres psql -U "$POSTGRES_USER" "$POSTGRES_DB"
```

## 9. 备注

- 本地开发仍然支持 SQLite
- Docker Compose 默认走 PostgreSQL
- 后端配置文件为 `config/config.docker.yaml`
- 运行时环境变量以 `ADLAB_DATABASE_*` 为准，可覆盖配置文件

## 10. 阿里云 / 腾讯云推荐顺序

如果你最终部署在阿里云或腾讯云，推荐按这个顺序操作：

1. 在云服务器安装 Docker 与 Docker Compose
2. 在 ACR 或 TCR 创建命名空间
3. 先把基础镜像同步到你自己的仓库
4. 在服务器上使用 `.env.aliyun.example` 或 `.env.tencent.example`
5. 把示例里的 `YOUR_NAMESPACE` 改成你的实际仓库地址
6. 先执行：

```bash
docker pull "$POSTGRES_IMAGE"
docker pull "$NGINX_IMAGE"
docker pull "$GO_BASE_IMAGE"
docker pull "$NODE_BASE_IMAGE"
docker pull "$BACKEND_RUNTIME_IMAGE"
```

7. 镜像都能拉下来后，再执行：

```bash
docker-compose up -d --build
docker-compose run --rm --profile tools seed
```

这样做的好处是，问题会被提前暴露在“镜像源是否可用”这一步，而不会拖到整个 Compose 启动时才发现。
