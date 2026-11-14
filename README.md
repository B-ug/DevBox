# DevBox

Windows 上的开发者工具箱。界面使用 [MyGo](https://mygo.egoist.dev) Native UI 绘制，不使用 WebView。

当前版本 `0.1.0`，优先支持 Windows 11。

## 已有功能

- 主窗口：默认 1200×800，最小 900×600
- 工具栏、侧边导航、工作区和状态栏
- 深色 / 浅色主题，默认深色；也可以跟随 Windows 外观
- 设置页：从侧边栏「应用」打开，在其中选择深色、浅色或跟随系统，再回到刚才的工具
- JSON 工具
  - 格式化、压缩、去除转义、校验
  - 清空、复制输出
  - 左右分栏编辑器，输出只读，支持 JSON 语法高亮
  - 输出已有结果时，后续操作处理输出；修改输入后，下一次操作改用新的输入
- Base64 工具：标准与 URL Safe 编码、解码；换行和空格会被忽略，缺少填充符也可以解码
- 时间戳工具：秒与毫秒由用户选择，不按位数判断；与日期互转；当前时间每秒刷新；同时显示本地和 UTC

## TODO

第一阶段的工具和打包已经完成。还剩下这些：

### 后续

- [ ] 用 MyGo 应用目录保存配置
- [ ] 最近使用、收藏工具
- [ ] JSON Diff

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
- `devbox-setup-0.1.0-amd64.exe`

## 目录

| 路径 | 内容 |
| --- | --- |
| `main.go` | 程序入口 |
| `internal/app` | 启动和主窗口 |
| `internal/ui` | 界面布局，以及 JSON、Base64、时间戳操作 |
| `internal/tools/json` | JSON 格式化、压缩、去转义和校验 |
| `internal/tools/base64` | Base64 标准与 URL Safe 编解码 |
| `internal/tools/datetime` | 时间戳与日期互转 |
| `internal/model` | 应用状态和工具列表 |
| `mygo.json` | 应用名称、标识和版本 |
