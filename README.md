# zvx - Kubernetes 容器管理平台

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.24.5-blue.svg)](https://go.dev/)
[![Vue Version](https://img.shields.io/badge/vue-3.5-green.svg)](https://vuejs.org/)

## 项目简介

zvx 是一个基于 **Vue.js 3** 和 **Go** 构建的现代化 Kubernetes 容器管理平台。它提供直观的 Web 界面，帮助用户管理 Kubernetes 集群中的容器、网络、存储和资源。

## 主要功能

- **容器管理**: 可视化的容器生命周期管理，包括创建、启动、停止、重启和删除
- **网络管理**: K8s 网络配置、虚拟网络设置和网络策略管理
- **存储管理**: 持久化存储卷、存储池和快照管理
- **资源监控**: CPU、内存、网络使用率的实时监控和图表展示
- **终端访问**: 直接在平台中访问容器终端，执行命令行
- **多语言支持**: 支持中文、英文、日文、韩文等多种语言
- **日志查询**: 实时查看和搜索容器日志
- **服务管理**: 负载均衡器、服务发现和虚拟 IP 配置
- **部署管理**: 容器镜像部署和管理

## 技术栈

### 前端 (zvx_web)
- **框架**: Vue.js 3 + Vite
- **UI 组件**: Element Plus
- **图表**: ECharts, Chart.js
- **代码编辑**: Monaco Editor, CodeMirror
- **终端**: XTerm.js
- **状态管理**: Pinia
- **国际化**: Vue i18n

### 后端 (zvx_go)
- **语言**: Go 1.24.5
- **Web 框架**: Gin
- **Kubernetes 客户端**: k8s.io/client-go
- **WebSocket**: Gorilla WebSocket
- **日志**: Custom logger
- **路由**: Gin Router

- **默认密码**: admin admin123

## 项目结构

```
zvx/
├── zvx_web/              # 前端 Vue.js 应用
│   ├── src/
│   │   ├── components/  # Vue 组件
│   │   │   ├── dashboard/  # 仪表盘组件
│   │   │   │   ├── PodManager.vue  # 容器管理
│   │   │   │   ├── PodNetwork.vue  # 网络配置
│   │   │   │   ├── StorageManager.vue  # 存储管理
│   │   │   │   ├── Status.vue  # 状态监控
│   │   │   │   ├── DeployImage.vue  # 镜像部署
│   │   │   │   ├── TerminalPage.vue  # 终端访问
│   │   │   │   ├── Settings.vue  # 设置
│   │   │   └── Pod/
│   │   │       └── PodCard.vue  # 容器卡片
│   │   ├── views/       # 页面视图
│   │   ├── locales/     # 多语言文件
│   │   ├── stores/      # Pinia 状态存储
│   │   ├── router/      # 路由配置
│   │   └── utils/       # 工具函数
│   ├── public/          # 静态资源
│   ├── login-go/        # Go 登录服务
│   └── login-linux       # Linux 登录脚本
│
├── zvx_go/              # 后端 Go API 服务
│   ├── controller/     # 控制器层
│   │   ├── pod.go       # 容器控制逻辑
│   │   ├── network.go   # 网络控制逻辑
│   │   ├── storage.go   # 存储控制逻辑
│   │   ├── service.go   # 服务控制逻辑
│   │   ├── terminal.go  # 终端控制逻辑
│   │   └── ...
│   ├── service/         # 服务层
│   ├── router/          # HTTP 路由
│   ├── kube/            # Kubernetes 客户端封装
│   ├── logs/            # 日志目录
│   ├── pkg/             # 工具包
│   └── storage/         # 存储配置
│
└── deploy.sh            # Docker 部署脚本
```

## 安装与部署

### 前置要求

- Docker 20.10+
- Docker Compose 1.29+
- 或 Kustomize Kubernetes 集群（可选）

### 快速开始

1. **克隆仓库**
```bash
git clone https://github.com/AutumnNode/zvx.git
cd zvx
```

2. **部署方式**

   **方式 1: 后端 Docker 部署**
   
   首先运行后端部署脚本构建镜像：
   ```bash
   cd zvx_go
   ./deploy.sh
   ```

   **方式 2: 前端 Docker 部署**
   
   构建并启动前端服务：
   ```bash
   cd zvx_web
   docker build -t zvx-frontend .
   docker run -d \
     -p 5173:5173 \
     -p 8080:8080 \
     zvx-frontend
   ```

   **方式 3: 本地开发部署**
   
   分别启动前端和后端：
   ```bash
   # 启动后端
   cd zvx_go
   go mod tidy && go run main.go &

   # 启动前端
   cd ../zvx_web
   npm install && npm run dev
   ```

3. **访问应用**
   - 前端界面：http://localhost:5173
   - 后端 API：http://localhost:8081

## API 文档

后端 API 运行在 `http://localhost:8081`，提供以下主要接口：

- `/api/pods` - 容器管理
- `/api/network` - 网络配置
- `/api/storage` - 存储管理
- `/api/service` - 服务管理
- `/api/terminal` - 终端访问
- `/api/logs` - 日志查询

## 开发

### 前端开发

```bash
cd zvx_web
npm install
npm run dev
```

### 后端开发

```bash
cd zvx_go
go mod tidy
go run main.go
```

### 构建

```bash
# 前端构建
cd zvx_web
npm run build

# 后端构建
cd zvx_go
go build
```

## 配置

### 环境变量

后端服务需要以下环境变量：

- `KUBECONFIG_PATH` - Kubernetes 配置文件路径（必需）
- `SERVER_PORT` - 服务器端口（默认：8081）
- `LOG_LEVEL` - 日志级别（默认：info）

### 前端配置

编辑 `zvx_web/src/vite-env.d.ts` 或 `vite.config.ts` 来配置代理设置。

## 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/YourAmazingFeature`)
3. 提交更改 (`git commit -m '添加 SomeFeature'`)
4. 推送到分支 (`git push origin feature/YourAmazingFeature`)
5. 创建 Pull Request

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 致谢

本项目使用了以下开源项目的代码：

- [Vue.js](https://vuejs.org/)
- [Element Plus](https://element-plus.org/)
- [ECharts](https://echarts.apache.org/)
- [Kubernetes Client Go](https://github.com/kubernetes/client-go)
- [Gin](https://github.com/gin-gonic/gin)

## 联系方式

- GitHub: [AutumnNode/zvx](https://github.com/AutumnNode/zvx)
- Issues: [报告问题](https://github.com/AutumnNode/zvx/issues)

---

**注意**: 本项目需要在 Kubernetes 集群环境中使用，请确保正确配置 kubeconfig 文件以连接到您的 K8s 集群。
