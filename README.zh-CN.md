# CLI Manager

一个跨平台桌面应用：从 GitHub Releases 安装和维护 CLI 工具，并在安装前验证官方 SHA-256 校验值。

[English](README.md)

## 主要能力

- 安装、更新、降级和卸载 CLI，实时显示进度。
- 下载后强制进行 SHA-256 校验，再替换目标文件。
- 探测已安装版本，并与最新正式版比较。
- 粘贴 GitHub 仓库地址添加工具，可自定义资产和校验文件命名规则。
- 可选接入 DeepSeek，根据仓库 README 生成中英双语说明。
- 界面支持跟随系统、简体中文和 English，并持久化选择。
- GitHub 和 DeepSeek token 只保存在本机，不会回传到前端。

## 内置工具

首次启动会加入 5 个已核实官方 Release 结构的工具：

| 工具 | 用途 | 平台 |
| --- | --- | --- |
| [asc](https://github.com/rorkai/App-Store-Connect-CLI) | App Store Connect 自动化 | macOS、Linux、Windows |
| [gh](https://github.com/cli/cli) | GitHub CLI | macOS、Linux、Windows |
| [jq](https://github.com/jqlang/jq) | JSON 处理 | macOS、Linux、Windows |
| [yq](https://github.com/mikefarah/yq) | YAML/JSON/XML 处理 | macOS、Linux、Windows |
| [fzf](https://github.com/junegunn/fzf) | 模糊搜索 | macOS、Linux、Windows |

本项目只支持同时发布可执行文件（或 ZIP/TAR.GZ）及兼容 SHA-256 校验文件的 Release。上游资产结构可能变化，届时需要在工具配置中调整模板。

## 开发环境

- Go 1.23+
- Node.js 20+ 与 npm
- [Wails v2.12](https://wails.io/docs/gettingstarted/installation)
- Wails 对应平台依赖（Windows 需要 WebView2）

```bash
git clone https://github.com/biyan113/cli-manager.git
cd cli-manager
wails dev
```

测试与构建：

```bash
go test ./...
cd frontend && npm ci && npm run build
wails build
```

产物位于 `build/bin/`。Windows NSIS 安装包使用：

```powershell
wails build -clean -platform windows/amd64 -nsis
```

## 配置目录

- Windows：`%AppData%\cli-manager\`
- macOS：`~/Library/Application Support/cli-manager/`
- Linux：`$XDG_CONFIG_HOME/cli-manager/` 或 `~/.config/cli-manager/`

Windows 默认安装到 `~/bin`，macOS/Linux 默认安装到 `~/.local/bin`。如命令无法直接运行，请将对应目录加入 `PATH`。

Windows self-hosted runner 的部署与发布流程见 [docs/windows-runner.md](docs/windows-runner.md)。只有真实 Windows runner 完成构建并通过产物冒烟验证后，才能视为 Windows 发布验证完成。

提交代码前请阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。安全问题请按 [SECURITY.md](SECURITY.md) 私下报告。

## 许可证

[MIT](LICENSE) © 2026 biyan
