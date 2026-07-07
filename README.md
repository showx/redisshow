# redisshow

基于 Go 编写的 Redis 命令行客户端，编译后生成单个可执行文件，可直接使用。支持 CLI 子命令、交互式 Shell 和 TUI 图形界面。

## 功能特性

- **CLI 命令**：常用 Redis 读写操作，适合脚本和快速查询
- **交互式 Shell**：轻量 REPL，适合临时操作
- **TUI 窗口**：左侧键列表、右侧详情，支持搜索高亮、删除、编辑、切换数据库

## 环境要求

- Go 1.21+
- 可访问的 Redis 服务

## 编译

```bash
go build -o redisshow .
```

Windows 下生成 `redisshow.exe`，Linux/macOS 下生成 `redisshow`。

## 连接配置

通过全局参数或环境变量配置 Redis 连接：

| 参数 | 环境变量 | 默认值 | 说明 |
|------|----------|--------|------|
| `-a, --addr` | `REDIS_ADDR` | `localhost:6379` | Redis 地址 |
| `-p, --password` | `REDIS_PASSWORD` | 空 | Redis 密码 |
| `-n, --db` | `REDIS_DB` | `0` | 数据库编号 |

示例：

```bash
# 命令行参数
./redisshow -a 127.0.0.1:6379 -p secret -n 1 ping

# 环境变量
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=secret
export REDIS_DB=0
./redisshow ping
```

## CLI 命令

```bash
./redisshow ping                          # 测试连接
./redisshow get mykey                     # 获取字符串
./redisshow set mykey "hello"             # 设置字符串
./redisshow set mykey "temp" --ttl 60s    # 设置并指定过期时间
./redisshow del key1 key2                 # 删除键
./redisshow keys "user:*"                 # 按模式列出键
./redisshow info                          # 查看服务信息
./redisshow type mykey                    # 查看键类型
./redisshow ttl mykey                     # 查看剩余 TTL（秒）
./redisshow hget user:1 name              # 获取哈希字段
./redisshow hset user:1 name alice        # 设置哈希字段
./redisshow lpush mylist a b c            # 列表左侧插入
./redisshow lrange mylist 0 -1            # 获取列表区间
./redisshow shell                         # 进入交互式 Shell
./redisshow tui                           # 打开 TUI 窗口
```

查看全部命令：

```bash
./redisshow --help
```

## TUI 模式（推荐）

```bash
./redisshow tui
```

### 界面说明

```
┌────────────────────────────────────────────────────────────┐
│ redisshow  localhost:6379  db=0  |  快捷键提示              │
├──────────────────┬───────────────────────────────────────────┤
│     键列表        │              键详情                       │
│  user:1          │  Key / Type / TTL / Value                 │
│  user:2          │                                           │
├──────────────────┴───────────────────────────────────────────┤
│ 匹配: user:*                                                │
│ 共 N 个键                                                   │
└────────────────────────────────────────────────────────────┘
```

- **左侧**：键列表，匹配模式中非通配符部分会黄色高亮
- **右侧**：显示键类型、TTL 和值（支持 string / hash / list / set / zset）
- **底部**：键名匹配模式，默认 `*`

### 快捷键

| 按键 | 功能 |
|------|------|
| `↑` `↓` | 切换键，右侧自动刷新详情 |
| `/` | 聚焦搜索框 |
| `Enter` | 确认搜索并刷新列表 |
| `r` | 刷新键列表 |
| `d` | 删除当前选中的键 |
| `e` | 编辑当前键（string / hash） |
| `b` | 切换数据库 |
| `q` / `Ctrl+C` | 退出 |

### 编辑说明

- **string**：打开文本编辑区，`Ctrl+S` 保存，`Esc` 取消
- **hash**：输入字段名和值，执行 `HSET`
- list / set / zset 暂不支持编辑

## 交互式 Shell

```bash
./redisshow shell
```

支持 `ping`、`get`、`set`、`del`、`keys`、`info`、`help`、`exit` 等命令。

## 项目结构

```
redisshow/
├── main.go
├── go.mod
├── internal/
│   ├── client/     # Redis 连接配置
│   ├── cmd/        # CLI 子命令
│   └── tui/        # TUI 界面
```

## 依赖

- [go-redis](https://github.com/redis/go-redis) — Redis 客户端
- [cobra](https://github.com/spf13/cobra) — CLI 框架
- [tview](https://github.com/rivo/tview) — TUI 界面
