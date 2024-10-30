# mac-vim-switch

一个为 Vim 用户设计的 macOS 输入法切换工具，基于 macism 开发。

## 功能特性

- 按下 ESC 键时自动切换到 ABC 输入法
- 使用 Shift 键在 ABC 和微信输入法拼音之间切换
- 以后台服务方式运行
- 易于与 macOS 系统集成

## 前置要求

- macOS 系统
- [macism](https://github.com/laishulu/macism)

## 安装方法

```bash
# 首先安装 macism
brew tap laishulu/macism
brew install macism

# 安装 mac-vim-switch
brew tap jackiexiao/mac-vim-switch
brew install mac-vim-switch

# 启动服务
brew services start mac-vim-switch
```

## 系统要求

1. 授予辅助功能权限
   - 前往 系统偏好设置 > 安全性与隐私 > 隐私 > 辅助功能
   - 点击锁图标以进行更改
   - 将 mac-vim-switch 添加到允许的应用列表中
   - 勾选 mac-vim-switch 旁边的复选框

2. 系统偏好设置 > 键盘 > 快捷键 > 输入法
   - 禁用"选择上一个输入法"
   - 禁用"选择输入菜单中的下一个输入法"

## 使用方法

安装后服务会自动启动并在后台运行。你可以：

- 按 ESC 键切换到 ABC 输入法
- 按 Shift 键在 ABC 和微信输入法拼音之间切换

### 服务管理

```bash
# 启动服务
brew services start mac-vim-switch

# 停止服务
brew services stop mac-vim-switch

# 重启服务
brew services restart mac-vim-switch

# 查看服务状态
brew services list
```

### 查看可用的输入法

如果你想使用其他输入法，可以查看可用的输入法 ID：

```bash
macism
```

### 日志

日志文件存储在 `~/.mac-vim-switch.log`

### 配置

你可以使用以下命令配置输入法：

```bash
# 列出可用的输入法
mac-vim-switch list

# 设置主输入法（默认：com.apple.keylayout.ABC）
mac-vim-switch config primary "com.apple.keylayout.ABC"

# 设置第二输入法
mac-vim-switch config secondary "your.input.method.id"
```

配置文件存储在 `~/.config/mac-vim-switch/config.json`

## 开发者指南

### 从源码构建

1. 克隆仓库
```bash
git clone https://github.com/jackiexiao/mac-vim-switch.git
cd mac-vim-switch
```

2. 安装依赖：
   - Go 1.16 或更高版本
   - macism (`brew install macism`)
   - Xcode Command Line Tools (用于 CGo 编译)

3. 构建项目：
```bash
go build ./cmd/mac-vim-switch
```

### 开发和调试

1. 在调试模式下运行并查看日志：
```bash
# 构建并本地运行
go build ./cmd/mac-vim-switch
./mac-vim-switch

# 实时查看日志
tail -f ~/.mac-vim-switch.log
```

2. 测试不同命令：
```bash
# 检查版本
./mac-vim-switch --version

# 列出可用输入法
./mac-vim-switch list

# 检查当前配置
./mac-vim-switch config

# 运行健康检查
./mac-vim-switch health
```

3. 调试 CGo 和键盘事件：
```bash
# 构建带调试符号的版本
go build -gcflags="all=-N -l" ./cmd/mac-vim-switch

# 使用详细的 CGo 日志运行
GODEBUG=cgocheck=2 ./mac-vim-switch

# 检查键盘事件是否被捕获
log stream --predicate 'process == "mac-vim-switch"'
```

4. 常见开发任务：
   - 修改输入法行为：编辑 main.go 中的 `switchToInputMethod()`
   - 添加新命令：在主命令 switch 中添加 case
   - 修改键盘处理：编辑 main.go 中的 CGo 回调
   - 更改默认设置：修改 main.go 顶部的常量值

5. 测试安装：
```bash
# 构建并本地安装
go build ./cmd/mac-vim-switch
sudo cp mac-vim-switch /usr/local/bin/

# 作为服务测试
brew services stop mac-vim-switch  # 如果正在运行则停止服务
./mac-vim-switch                  # 直接运行以查看日志
```

### 开发问题排查

1. CGo 编译错误：
   - 确保已安装 Xcode Command Line Tools：`xcode-select --install`
   - 检查 main.go 中的 CGo 标志
   - 尝试清理构建：`go clean -cache`

2. 键盘事件问题：
   - 检查系统偏好设置中的辅助功能权限
   - 使用 sudo 测试权限：`sudo ./mac-vim-switch`
   - 启用调试日志：`GODEBUG=cgocheck=2 ./mac-vim-switch`

3. 输入法切换问题：
   - 直接测试 macism：`macism "com.apple.keylayout.ABC"`
   - 检查可用方法：`macism`
   - 验证权限：`mac-vim-switch health`

### 项目结构

- `cmd/mac-vim-switch/main.go`：主程序
  - 键盘事件的 CGo 绑定
  - 输入法切换逻辑
  - 配置管理
- `Formula/mac-vim-switch.rb`：Homebrew 配方
- `mac-vim-switch.plist`：LaunchAgent 配置
- `.config/mac-vim-switch/config.json`：运行时配置

### 修改代码

1. 更新版本号：
   - `main.go`（`version` 常量）
   - `Formula/mac-vim-switch.rb`

2. 测试更改：
   - 本地构建和运行
   - 检查日志中的错误
   - 验证键盘事件
   - 测试配置更改

3. 提交前：
   - 运行 `go fmt ./...`
   - 运行 `go vet ./...`
   - 在干净的系统上测试

## 故障排除

1. 如果服务不工作：
   - 检查是否已授予辅助功能权限
   - 查看日志：`cat ~/.mac-vim-switch.log`
   - 尝试重启服务：`brew services restart mac-vim-switch`

2. 如果输入法切换不工作：
   - 运行 `macism` 检查可用的输入法
   - 确保输入法已正确安装

## 许可证

MIT 许可证