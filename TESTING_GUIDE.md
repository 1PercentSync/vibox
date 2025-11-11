# ViBox 完整测试指南

> **目的**: 本指南将帮助你在本地环境中完整测试ViBox的所有功能
>
> **测试日期**: 2025-11-11
> **版本**: v1.0.0

---

## 目录

1. [环境准备](#环境准备)
2. [启动ViBox服务](#启动vibox服务)
3. [Web界面测试](#web界面测试)
4. [创建Ubuntu工作空间](#创建ubuntu工作空间)
5. [测试WebSSH终端](#测试webssh终端)
6. [测试端口映射和HTTP服务](#测试端口映射和http服务)
7. [安装code-server](#安装code-server)
8. [故障排查](#故障排查)

---

## 环境准备

### 前置条件

- Docker 20.10+
- Docker Compose v2.0+
- 浏览器（Chrome/Firefox/Safari）

### 检查Docker状态

```bash
# 检查Docker版本
docker --version

# 确认Docker运行中
docker info

# 检查Docker Compose
docker compose version
```

---

## 启动ViBox服务

### 1. 克隆并配置

```bash
# 进入项目目录
cd /path/to/vibox

# 检查是否有.env文件
ls -la .env

# 如果没有，创建.env文件
echo "API_TOKEN=$(openssl rand -hex 32)" > .env
```

### 2. 启动服务

```bash
# 构建并启动ViBox服务
docker compose up -d --build

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f vibox
```

**预期输出**:
```
NAME      IMAGE                               COMMAND        SERVICE   CREATED         STATUS                   PORTS
vibox     ghcr.io/1percentsync/vibox:latest   "/app/vibox"   vibox     X seconds ago   Up X seconds (healthy)   0.0.0.0:3000->3000/tcp
```

### 3. 验证服务健康

```bash
# 健康检查
curl http://localhost:3000/health

# 预期输出: {"service":"vibox","status":"ok"}
```

---

## Web界面测试

### 1. 访问前端界面

在浏览器中打开: **http://localhost:3000**

**预期显示**: ViBox登录页面

### 2. 登录

1. 获取API Token:
   ```bash
   cat .env | grep API_TOKEN
   ```

2. 在登录页面输入Token并点击"Login"

**预期结果**: 成功登录，跳转到工作空间列表页面

---

## 创建Ubuntu工作空间

### 方法1: 通过Web界面（推荐）

1. 点击 **"+ Create Workspace"** 按钮
2. 填写工作空间信息：
   - **Name**: `my-dev-workspace`
   - **Image**: `ubuntu:rolling`
   - **Scripts**: 点击 "Add Script" 添加以下脚本：

**脚本1 - 安装基础工具**:
```bash
apt-get update && apt-get install -y \
  curl \
  wget \
  git \
  vim \
  python3 \
  python3-pip
```

**脚本2 - 创建HTML测试文件**:
```bash
echo "<html><body><h1>Hello from ViBox!</h1><p>Test HTML service</p></body></html>" > /root/index.html
```

**脚本3 - 启动HTTP服务器**:
```bash
cd /root && nohup python3 -m http.server 8000 > /tmp/http.log 2>&1 &
```

3. 添加端口标签：
   - Port: `8000`
   - Label: `HTTP Server`

4. 点击 **"Create"**

**预期结果**:
- 工作空间状态显示为 "creating"
- 等待30-60秒后，状态变为 "running"

### 方法2: 通过API测试

```bash
# 获取Token
TOKEN=$(cat .env | grep API_TOKEN | cut -d= -f2)

# 登录获取Cookie
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -c /tmp/vibox-cookies.txt \
  -d "{\"token\": \"$TOKEN\"}"

# 创建工作空间
curl -X POST http://localhost:3000/api/workspaces \
  -b /tmp/vibox-cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-workspace",
    "image": "ubuntu:rolling",
    "scripts": [
      {
        "name": "install-tools",
        "content": "apt-get update && apt-get install -y python3 curl vim",
        "order": 0
      },
      {
        "name": "create-html",
        "content": "echo \"<html><body><h1>Hello from ViBox!</h1></body></html>\" > /root/index.html",
        "order": 1
      },
      {
        "name": "start-http",
        "content": "cd /root && nohup python3 -m http.server 8000 > /tmp/http.log 2>&1 &",
        "order": 2
      }
    ],
    "ports": {
      "8000": "HTTP Server"
    }
  }'

# 记录返回的workspace ID（例如: ws-xxxxxx）
```

### 验证工作空间创建

```bash
# 查看工作空间列表
curl -s http://localhost:3000/api/workspaces -b /tmp/vibox-cookies.txt

# 查看容器
docker ps --filter "label=vibox.workspace"

# 查看脚本执行日志
WORKSPACE_ID="ws-xxxxxx"  # 替换为实际ID
docker exec vibox-$WORKSPACE_ID cat /var/log/vibox/install-tools.log
```

---

## 测试WebSSH终端

### 通过Web界面

1. 在工作空间列表中，找到刚创建的工作空间
2. 点击 **"Terminal"** 按钮
3. 等待终端加载

**预期结果**:
- 显示黑色终端窗口
- 看到命令提示符: `# `
- 可以输入命令

### 测试终端命令

在终端中执行以下命令：

```bash
# 检查系统信息
uname -a

# 查看Python版本
python3 --version

# 检查HTTP服务
ps aux | grep python

# 查看HTTP服务日志
cat /tmp/http.log

# 测试网络连接
curl ifconfig.me
```

### 通过命令行测试WebSocket

```bash
# 使用websocat测试（需要安装websocat）
TOKEN=$(cat .env | grep API_TOKEN | cut -d= -f2)
WORKSPACE_ID="ws-xxxxxx"  # 替换为实际ID

timeout 5 websocat "ws://localhost:3000/ws/terminal/$WORKSPACE_ID?token=$TOKEN"
```

**预期输出**: 看到终端提示符

---

## 测试端口映射和HTTP服务

### 1. 通过Web界面访问

1. 在工作空间详情页，点击 **"Ports"** 标签
2. 找到 "HTTP Server (8000)" 条目
3. 点击 **"Open"** 按钮

**预期结果**:
- 在新窗口中打开HTTP服务
- 显示HTML页面: "Hello from ViBox!"

### 2. 通过API测试端口转发

```bash
# 访问端口转发URL
WORKSPACE_ID="ws-xxxxxx"  # 替换为实际ID
curl http://localhost:3000/forward/$WORKSPACE_ID/8000/ -b /tmp/vibox-cookies.txt

# 预期输出: <html><body><h1>Hello from ViBox!</h1></body></html>
```

### 3. 测试任意端口访问

在终端中启动另一个服务：

```bash
# 在WebSSH终端中执行
python3 -m http.server 9000 &

# 等待几秒后，通过浏览器访问
# http://localhost:3000/forward/ws-xxxxxx/9000/
```

---

## 安装code-server

### 1. 通过WebSSH终端安装

在工作空间的WebSSH终端中执行：

```bash
# 安装code-server
curl -fsSL https://code-server.dev/install.sh | sh

# 配置code-server（无密码模式，仅用于测试）
mkdir -p ~/.config/code-server
cat > ~/.config/code-server/config.yaml <<EOF
bind-addr: 0.0.0.0:8080
auth: none
cert: false
EOF

# 启动code-server
nohup code-server > /tmp/code-server.log 2>&1 &

# 等待几秒让服务启动
sleep 5

# 检查code-server是否运行
ps aux | grep code-server
cat /tmp/code-server.log
```

### 2. 添加code-server端口标签

#### 通过Web界面:
1. 在工作空间详情页，点击 "Ports" 标签
2. 点击 **"+ Add Port"**
3. 输入:
   - Port: `8080`
   - Label: `VS Code Server`
4. 点击 "Add"
5. 点击新添加的端口的 "Open" 按钮

#### 通过API:
```bash
WORKSPACE_ID="ws-xxxxxx"  # 替换为实际ID

curl -X PUT http://localhost:3000/api/workspaces/$WORKSPACE_ID/ports \
  -b /tmp/vibox-cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "ports": {
      "8000": "HTTP Server",
      "8080": "VS Code Server"
    }
  }'
```

### 3. 访问code-server

在浏览器中打开:
```
http://localhost:3000/forward/<WORKSPACE_ID>/8080/
```

**预期结果**:
- 显示VS Code Web界面
- 可以正常编辑文件
- 可以打开终端

### 4. 测试code-server功能

在VS Code界面中：

1. **创建文件**:
   - 点击 File → New File
   - 创建 `hello.py`:
     ```python
     print("Hello from ViBox + code-server!")
     ```

2. **运行代码**:
   - 打开终端 (Terminal → New Terminal)
   - 执行: `python3 hello.py`

3. **测试Git**:
   ```bash
   git config --global user.name "Test User"
   git config --global user.email "test@example.com"
   git init test-repo
   cd test-repo
   echo "# Test" > README.md
   git add README.md
   git commit -m "Initial commit"
   ```

---

## 故障排查

### 问题1: 服务无法启动

**症状**: `docker compose up` 失败

**排查步骤**:
```bash
# 检查环境变量
cat .env

# 检查Docker日志
docker compose logs vibox

# 检查端口占用
sudo lsof -i :3000

# 检查Docker socket权限
ls -la /var/run/docker.sock
```

**解决方案**:
```bash
# 确保API_TOKEN已设置
echo "API_TOKEN=$(openssl rand -hex 32)" > .env

# 如果端口被占用，修改端口
echo "HOST_PORT=3001" >> .env
docker compose up -d
```

### 问题2: 脚本未执行

**症状**: 工作空间创建成功但工具未安装

**排查步骤**:
```bash
# 检查脚本日志
WORKSPACE_ID="ws-xxxxxx"
docker exec vibox-$WORKSPACE_ID ls /var/log/vibox/
docker exec vibox-$WORKSPACE_ID cat /var/log/vibox/install-tools.log
```

**常见原因**:
- JSON中使用了 `"script"` 而不是 `"content"` 字段
- 脚本语法错误
- 网络问题导致apt-get失败

**解决方案**:
```bash
# 确保使用正确的字段名
{
  "scripts": [
    {
      "name": "test",
      "content": "echo 'test'",  // 使用 "content" 不是 "script"
      "order": 0
    }
  ]
}

# 重置工作空间（通过API）
curl -X POST http://localhost:3000/api/workspaces/$WORKSPACE_ID/reset \
  -b /tmp/vibox-cookies.txt
```

### 问题3: 端口转发无法访问

**症状**: 访问 `/forward/...` 返回 502 或超时

**排查步骤**:
```bash
# 检查容器内服务是否运行
docker exec vibox-$WORKSPACE_ID netstat -tlnp

# 检查HTTP服务
docker exec vibox-$WORKSPACE_ID curl localhost:8000

# 检查ViBox日志
docker compose logs vibox | tail -50
```

**解决方案**:
```bash
# 在容器内重启服务
docker exec -it vibox-$WORKSPACE_ID bash
cd /root && python3 -m http.server 8000 &
exit

# 或通过WebSSH终端重启服务
```

### 问题4: WebSocket连接失败

**症状**: 终端无法连接或立即断开

**排查步骤**:
```bash
# 检查工作空间状态
curl http://localhost:3000/api/workspaces/$WORKSPACE_ID -b /tmp/vibox-cookies.txt

# 检查容器是否运行
docker ps | grep vibox-$WORKSPACE_ID

# 检查浏览器控制台错误
# 按F12打开开发者工具，查看Network和Console标签
```

**解决方案**:
- 确保使用正确的Token
- 确保容器状态为 "running"
- 如果使用反向代理，确保WebSocket升级配置正确

### 问题5: code-server无法访问

**症状**: 访问8080端口返回404或空白页

**排查步骤**:
```bash
# 检查code-server进程
docker exec vibox-$WORKSPACE_ID ps aux | grep code-server

# 检查code-server日志
docker exec vibox-$WORKSPACE_ID cat /tmp/code-server.log

# 检查端口监听
docker exec vibox-$WORKSPACE_ID netstat -tlnp | grep 8080
```

**解决方案**:
```bash
# 重启code-server
docker exec -it vibox-$WORKSPACE_ID bash
pkill code-server
nohup code-server > /tmp/code-server.log 2>&1 &
exit

# 检查配置文件
docker exec vibox-$WORKSPACE_ID cat ~/.config/code-server/config.yaml
```

---

## 完整测试检查清单

使用此清单确保所有功能正常：

- [ ] ViBox服务启动成功
- [ ] 健康检查通过 (`/health`)
- [ ] Web界面可访问 (http://localhost:3000)
- [ ] 登录功能正常
- [ ] 可以创建工作空间
- [ ] 脚本自动执行成功
- [ ] WebSSH终端可连接
- [ ] 终端命令执行正常
- [ ] 端口转发功能正常
- [ ] HTTP服务可访问
- [ ] 可以添加端口标签
- [ ] code-server安装成功
- [ ] code-server可通过端口转发访问
- [ ] code-server编辑功能正常
- [ ] 工作空间删除功能正常

---

## 测试总结

### 成功的测试场景

测试日期: 2025-11-11

| 功能 | 状态 | 备注 |
|------|------|------|
| 前后端集成 | ✅ | 前端静态文件嵌入成功 |
| Docker构建 | ✅ | 多阶段构建工作正常 |
| 服务启动 | ✅ | 健康检查通过 |
| API认证 | ✅ | Cookie鉴权正常 |
| 工作空间创建 | ✅ | ubuntu:rolling容器创建成功 |
| 脚本执行 | ✅ | 3个初始化脚本成功执行 |
| WebSSH终端 | ✅ | WebSocket连接正常 |
| 端口转发 | ✅ | HTTP服务(8000)可访问 |
| 端口标签 | ✅ | 动态添加端口标签成功 |

### 已知问题

1. **脚本字段名混淆**:
   - 问题: API请求中使用 `"script"` 而非 `"content"` 导致脚本为空
   - 影响: 脚本不执行，但不报错
   - 解决: 文档中明确说明使用 `"content"` 字段

2. **docker-compose.yml版本警告**:
   - 问题: `version` 字段已过时
   - 影响: 仅警告，不影响功能
   - 解决: 可选，删除 `version: '3.8'` 行

### 下一步建议

1. 编写前端自动化测试
2. 添加API集成测试套件
3. 性能测试（多个并发工作空间）
4. 安全测试（Token验证、容器隔离）
5. 添加监控和日志聚合

---

**测试完成！ViBox所有核心功能运行正常。**

服务现已运行在 http://localhost:3000

API Token可在 `.env` 文件中查看。
