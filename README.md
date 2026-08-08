# cli-manager

一个通用的 GitHub Releases CLI 工具管理器(Wails v2 + Go + React 桌面应用)。

## 功能

- **添加工具**：粘贴 GitHub 地址自动识别仓库、二进制名与资产命名模板，也可手动微调字段
- **一键安装 / 更新 / 降级 / 卸载**：从 GitHub Releases 解析资产 → 拉取 checksums → SHA256 校验 → 原子安装到 `~/.local/bin`，实时显示下载进度
- **版本对比**：自动探测已装版本与最新版本，标记「有新版本 / 已是最新 / 未安装」
- **工具说明**：接入 DeepSeek，为每个工具生成中英双语简介，并逐条列出最近版本更新说明
- **缓存**：最新版本查询 5 分钟 TTL、工具说明 1 小时 TTL，避免重复请求

## 技术栈

- **桌面框架**：Wails v2.12.0
- **后端**：Go（`net/http` 调用 GitHub Releases API、DeepSeek API）
- **前端**：React（Vite + JSX）

## 快速开始

```bash
# 开发模式（前端热更新）
wails dev

# 打包
wails build
```

产物位于 `build/bin/cli-manager.app`。

## 目录结构

```
cli-manager/
├── main.go              # 入口：embed frontend/dist，绑定 App
├── app.go               # Wails 绑定层：全部对外方法
├── internal/
│   ├── config/          # 配置读写（tools.json / state.json）
│   ├── github/          # GitHub Releases API 客户端
│   ├── deepseek/        # DeepSeek 客户端（工具说明）
│   └── tool/            # 安装编排：资产命名、checksums、下载、原子安装
└── frontend/            # React 前端
```

配置默认存放于 `~/Library/Application Support/cli-manager/`（macOS）。
