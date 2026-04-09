# jumpd

Windows CLI 目录快速跳转工具，基于模糊匹配快速定位目标目录。选择后自动打开新 CMD 窗口，零配置。

## 安装

```bash
go install github.com/scm1219/jumpd@latest
```

或从源码编译：

```bash
git clone https://github.com/scm1219/jumpd.git
cd jumpd
go build -o jumpd.exe .
```

将 `jumpd.exe` 放入 PATH 目录即可。

## 用法

```
jumpd <drive> [pattern1] [pattern2] ...
jumpd [flags] <drive> [pattern1] [pattern2] ...
```

- 第一个参数为盘符（支持 `d` 或 `d:` 格式）
- 后续参数为目录名模糊匹配词，按层级逐级匹配（不区分大小写）
- `-e` 可出现在任意位置

### 示例

```bash
# 搜索 D 盘下名称包含 "tools" 的目录，选择后打开新 CMD 窗口
jumpd d tools

# 使用 -e，选择后用 Windows 资源管理器打开目录（两种写法等价）
jumpd -e d tools
jumpd d tools -e

# 多级匹配：D 盘下 tools 目录中包含 "pickyou" 的子目录
jumpd d tools pickyou
```

### 交互控制

| 按键 | 功能 |
|------|------|
| `← →` | 翻页（左/右箭头） |
| `1-9` | 选择当前页目录 |
| `g` + 页码 + `Enter` | 跳转到指定页（如 `g3` 跳到第3页） |
| `Esc` | 取消跳转输入 |
| `q` | 退出 |

## 参数说明

```
Flags:
  -e, --explorer   用 Windows 资源管理器打开选中目录（默认打开新 CMD 窗口）
  -h, --help       显示帮助信息
```

## 许可

MIT
