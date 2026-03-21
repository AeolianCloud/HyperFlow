# API规范文档整合和发布指南

## 文档站点结构

```
docs-site/
├── index.html                    # 首页
├── getting-started/              # 快速开始
│   ├── index.md
│   └── authentication.md
├── design-standards/             # 设计标准
│   └── index.md
├── implementation-guidelines/    # 实现指南
│   └── index.md
├── documentation-standards/      # 文档规范
│   └── index.md
├── api-reference/                # API参考
│   ├── v1/
│   │   ├── openapi.yaml
│   │   └── index.html
│   └── v2/
│       ├── openapi.yaml
│       └── index.html
├── examples/                     # 示例代码
│   ├── go/
│   ├── javascript/
│   └── python/
├── tools/                        # 工具配置
│   └── index.md
└── changelog.md                  # 变更日志
```

## 1. 整合所有规范文档

### 创建文档站点首页

**index.html:**
```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API规范文档</title>
    <link rel="stylesheet" href="styles.css">
</head>
<body>
    <header>
        <h1>API规范文档</h1>
        <nav>
            <a href="#getting-started">快速开始</a>
            <a href="#standards">规范标准</a>
            <a href="#api-reference">API参考</a>
            <a href="#examples">示例代码</a>
        </nav>
    </header>

    <main>
        <section id="getting-started">
            <h2>快速开始</h2>
            <div class="cards">
                <div class="card">
                    <h3>认证</h3>
                    <p>了解如何获取和使用访问令牌</p>
                    <a href="getting-started/authentication.html">查看详情 →</a>
                </div>
                <div class="card">
                    <h3>第一个请求</h3>
                    <p>发送你的第一个API请求</p>
                    <a href="getting-started/first-request.html">查看详情 →</a>
                </div>
            </div>
        </section>

        <section id="standards">
            <h2>规范标准</h2>
            <div class="cards">
                <div class="card">
                    <h3>设计标准</h3>
                    <p>RESTful API设计规范</p>
                    <a href="design-standards/">查看文档 →</a>
                </div>
                <div class="card">
                    <h3>实现指南</h3>
                    <p>API实现最佳实践</p>
                    <a href="implementation-guidelines/">查看文档 →</a>
                </div>
                <div class="card">
                    <h3>文档规范</h3>
                    <p>API文档编写标准</p>
                    <a href="documentation-standards/">查看文档 →</a>
                </div>
            </div>
        </section>

        <section id="api-reference">
            <h2>API参考</h2>
            <div class="version-selector">
                <a href="api-reference/v2/" class="version active">v2 (最新)</a>
                <a href="api-reference/v1/" class="version">v1</a>
            </div>
        </section>

        <section id="examples">
            <h2>示例代码</h2>
            <div class="cards">
                <div class="card">
                    <h3>Go</h3>
                    <p>Go语言客户端示例</p>
                    <a href="examples/go/">查看示例 →</a>
                </div>
                <div class="card">
                    <h3>JavaScript</h3>
                    <p>JavaScript/TypeScript示例</p>
                    <a href="examples/javascript/">查看示例 →</a>
                </div>
                <div class="card">
                    <h3>Python</h3>
                    <p>Python客户端示例</p>
                    <a href="examples/python/">查看示例 →</a>
                </div>
            </div>
        </section>
    </main>

    <footer>
        <p>&copy; 2024 API规范文档. All rights reserved.</p>
    </footer>
</body>
</html>
```

## 2. 配置文档站点导航和搜索

### 使用Docsify

**安装和配置：**
```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>API规范文档</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <link rel="stylesheet" href="//cdn.jsdelivr.net/npm/docsify@4/lib/themes/vue.css">
</head>
<body>
  <div id="app"></div>
  <script>
    window.$docsify = {
      name: 'API规范文档',
      repo: 'https://github.com/org/api-docs',
      loadSidebar: true,
      subMaxLevel: 3,
      search: {
        placeholder: '搜索',
        noData: '没有结果',
        depth: 6
      },
      pagination: {
        previousText: '上一页',
        nextText: '下一页'
      }
    }
  </script>
  <script src="//cdn.jsdelivr.net/npm/docsify@4"></script>
  <script src="//cdn.jsdelivr.net/npm/docsify/lib/plugins/search.min.js"></script>
</body>
</html>
```

**_sidebar.md:**
```markdown
* 快速开始
  * [认证](getting-started/authentication.md)
  * [第一个请求](getting-started/first-request.md)

* 规范标准
  * [设计标准](design-standards/index.md)
  * [实现指南](implementation-guidelines/index.md)
  * [文档规范](documentation-standards/index.md)

* API参考
  * [v2 (最新)](api-reference/v2/)
  * [v1](api-reference/v1/)

* 示例代码
  * [Go](examples/go/)
  * [JavaScript](examples/javascript/)
  * [Python](examples/python/)

* 工具和配置
  * [工具配置](tools/index.md)
  * [审查流程](review-process.md)

* [变更日志](changelog.md)
```

## 3. 添加文档版本控制

### Git标签管理

```bash
# 为文档创建版本标签
git tag -a docs-v1.0.0 -m "API规范文档 v1.0.0"
git push origin docs-v1.0.0

# 查看所有文档版本
git tag -l "docs-v*"
```

### 版本切换器

```html
<div class="version-switcher">
  <select onchange="window.location.href=this.value">
    <option value="/docs/v2.0.0/" selected>v2.0.0 (最新)</option>
    <option value="/docs/v1.0.0/">v1.0.0</option>
  </select>
</div>
```

### 版本归档

```
docs/
├── latest/          -> v2.0.0/
├── v2.0.0/
│   ├── index.html
│   └── ...
├── v1.0.0/
│   ├── index.html
│   └── ...
└── versions.json
```

**versions.json:**
```json
{
  "latest": "v2.0.0",
  "versions": [
    {
      "version": "v2.0.0",
      "date": "2024-03-19",
      "status": "current"
    },
    {
      "version": "v1.0.0",
      "date": "2024-01-01",
      "status": "archived"
    }
  ]
}
```

## 4. 配置文档访问权限

### Nginx配置（公开访问）

```nginx
server {
    listen 80;
    server_name docs.example.com;

    root /var/www/docs;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # 启用gzip
    gzip on;
    gzip_types text/plain text/css application/json application/javascript;

    # 缓存静态资源
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### 基本认证（内部访问）

```nginx
server {
    listen 80;
    server_name internal-docs.example.com;

    root /var/www/docs;
    index index.html;

    auth_basic "API Documentation";
    auth_basic_user_file /etc/nginx/.htpasswd;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

**创建密码文件：**
```bash
htpasswd -c /etc/nginx/.htpasswd username
```

## 5. 发布文档并通知团队

### 自动化部署

**GitHub Actions部署到GitHub Pages：**
```yaml
name: Deploy Docs

on:
  push:
    branches: [main]
    paths:
      - 'docs/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build docs
        run: |
          # 复制所有文档到部署目录
          mkdir -p deploy
          cp -r docs/* deploy/

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./deploy
          cname: docs.example.com
```

### 发布通知

**Slack通知：**
```yaml
- name: Notify Slack
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    text: |
      📚 API文档已更新
      版本: v2.0.0
      查看: https://docs.example.com
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

**邮件通知模板：**
```markdown
主题: API规范文档已更新 - v2.0.0

团队成员，

API规范文档已更新到v2.0.0版本。

## 主要变更
- 新增分页机制规范
- 更新错误处理指南
- 添加Go语言示例代码

## 访问文档
https://docs.example.com

## 重要提醒
请所有开发人员在开发新API时参考最新规范。

如有问题，请联系API团队。

---
API团队
2024-03-19
```

### 发布检查清单

- [ ] 所有文档已生成
- [ ] 文档链接正常
- [ ] 搜索功能正常
- [ ] 版本切换正常
- [ ] 移动端显示正常
- [ ] 示例代码可运行
- [ ] 文档已部署到生产环境
- [ ] DNS配置正确
- [ ] SSL证书有效
- [ ] 已通知团队

### 发布后监控

```bash
# 检查文档可访问性
curl -I https://docs.example.com

# 检查搜索功能
curl https://docs.example.com/search?q=authentication

# 监控访问日志
tail -f /var/log/nginx/docs-access.log
```

## 完整部署脚本

```bash
#!/bin/bash
# deploy-docs.sh

set -e

echo "🚀 开始部署API文档..."

# 1. 生成API文档
echo "📝 生成API文档..."
swag init

# 2. 验证文档
echo "✅ 验证文档..."
spectral lint docs/swagger.yaml

# 3. 构建文档站点
echo "🔨 构建文档站点..."
mkdir -p deploy
cp -r docs/* deploy/
cp -r templates deploy/
cp -r examples deploy/

# 4. 部署到服务器
echo "📤 部署到服务器..."
rsync -avz --delete deploy/ user@docs-server:/var/www/docs/

# 5. 重启Nginx
echo "🔄 重启Nginx..."
ssh user@docs-server "sudo systemctl reload nginx"

# 6. 验证部署
echo "🔍 验证部署..."
curl -f https://docs.example.com || exit 1

# 7. 发送通知
echo "📢 发送通知..."
curl -X POST $SLACK_WEBHOOK \
  -H 'Content-Type: application/json' \
  -d '{"text":"📚 API文档已更新: https://docs.example.com"}'

echo "✨ 部署完成！"
```

## 使用方式

```bash
# 赋予执行权限
chmod +x deploy-docs.sh

# 执行部署
./deploy-docs.sh
```
