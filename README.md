# DevBox

Windows 上的开发者工具箱。界面使用 [MyGo](https://mygo.egoist.dev) Native UI 绘制，不使用 WebView。

优先支持 Windows 11。

## 已有功能

- 主窗口：默认 1200×800，最小 900×600
- 工具栏、侧边导航、工作区和状态栏
- 深色 / 浅色主题，默认深色；也可以跟随 Windows 外观
- 设置页：从侧边栏「应用」打开，在其中选择深色、浅色或跟随系统，并配置翻译用的模型，再回到刚才的工具
- 主题保存在 MyGo 应用目录的 `config.json`，下次启动会沿用。Windows 上，`go run` 和安装后的程序使用 `%AppData%\DevBox\config.json`；`mygo dev` 使用 `%AppData%\DevBox Dev\config.json`
- 最近使用：打开过的工具出现在侧边栏，最近的在最上面，最多 8 个
- 收藏：工具页的「收藏」把当前工具放进侧边栏「收藏工具」，再点一次取消
- 最近使用、收藏、主题和翻译设置一起写入 `config.json`
- 检查更新：关于页的「检查更新」打开 MyGo 的更新窗口。安装后的程序会在启动后检查 [GitHub Releases](https://github.com/B-ug/DevBox/releases)，有新版本时可以下载、安装并重新启动。`go run` 和 `mygo dev` 会说明当前是开发版本，不能自行更新
- JSON 工具
  - 格式化、压缩、去除转义、校验
  - 清空、复制当前内容
  - 单个编辑器，操作直接改写其中的文本，支持 JSON 语法高亮；失败时保留原文
- Base64 工具：标准与 URL Safe 编码、解码，结果写回同一个编辑器；换行和空格会被忽略，缺少填充符也可以解码
- 时间戳工具：秒与毫秒由用户选择，不按位数判断；与日期互转；当前时间每秒刷新；同时显示本地和 UTC
- PDF 工具：按 DPI 和质量缩小内嵌图片并重新编码为 JPEG，文字、矢量和注释保留；输出为 `原文件名_compressed.pdf`
- 翻译工具：原文和译文分开放，选择目标语言后调用模型。模型在设置页配置
  - Base URL、API Key、模型名和 Prompt 在设置页修改，并写入本机 `config.json`。API Key 用密码框遮住
  - `{{target}}` 会替换成目标语言，原文作为用户消息发送
  - 接口是 OpenAI 兼容的 Chat Completions。默认地址 `https://api.openai.com/v1`，默认模型 `gpt-4o-mini`

## 环境

- Go 1.27 或更高版本
- 不需要 Bun，也不需要 WebView Runtime

## 开发

```powershell
go test ./...          # 业务测试和无窗口界面测试
go run .               # 启动当前程序
go tool mygo dev       # 热重载开发
go tool mygo build     # 打包到 build/
```

这次构建的产物在 `build/windows-amd64/`：

- `DevBox.exe`
- 安装包 `*-setup-*.exe`

检查更新的签名密钥这样生成。把命令打印的公钥填进 `mygo.json` 的 `updates.publicKey`，私钥内容设为仓库 Secret `MYGO_UPDATER_PRIVATE_KEY`。私钥不要提交到仓库。

```powershell
go tool mygo keygen
```

推送与 `mygo.json` 版本一致的 `v*` 标签后，工作流会把安装包和签名更新上传到草稿 Release；发布该草稿后，已安装的程序才能看到这一版。

## 目录

| 路径 | 内容 |
| --- | --- |
| `main.go` | 程序入口 |
| `internal/app` | 启动和主窗口 |
| `internal/ui` | 界面布局，以及 JSON、Base64、时间戳、PDF、翻译操作 |
| `internal/tools/json` | JSON 格式化、压缩、去转义和校验 |
| `internal/tools/base64` | Base64 标准与 URL Safe 编解码 |
| `internal/tools/datetime` | 时间戳与日期互转 |
| `internal/tools/pdf` | 缩小 PDF 内嵌图片并写回 |
| `internal/tools/translate` | 调用 OpenAI 兼容接口翻译文本 |
| `internal/model` | 应用状态和工具列表 |
| `internal/config` | 读写 MyGo 应用目录里的 `config.json` |
| `mygo.json` | 应用名称、标识和版本 |
