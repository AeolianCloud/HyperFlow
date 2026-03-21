# API规范工具配置指南

## 1. OpenAPI文档生成工具配置

### 使用swag（Go语言推荐）

**安装：**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

**配置文件 `.swaggo`：**
```yaml
# Swag配置
searchDir: ./
excludes: ./vendor,./docs
parseVendor: false
parseDependency: false
parseInternal: false
parseDepth: 100
```

**生成文档：**
```bash
swag init
```

**集成到项目：**
```go
import (
    _ "myapp/docs"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

## 2. Swagger UI配置

### 基本配置

**创建 `swagger-ui/index.html`：**
```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>API文档</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="./config.js"></script>
</body>
</html>
```

**创建 `swagger-ui/config.js`：**
```javascript
window.onload = function() {
  SwaggerUIBundle({
    url: "/openapi.yaml",
    dom_id: '#swagger-ui',
    deepLinking: true,
    displayRequestDuration: true,
    filter: true,
    persistAuthorization: true,
    docExpansion: "list",
    defaultModelsExpandDepth: 1,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIBundle.SwaggerUIStandalonePreset
    ],
    layout: "BaseLayout"
  });
};
```

## 3. 文档自动生成流程配置

### Makefile配置

```makefile
# Makefile
.PHONY: docs docs-serve docs-validate

# 生成API文档
docs:
	@echo "Generating API documentation..."
	swag init
	@echo "Documentation generated successfully"

# 本地预览文档
docs-serve:
	@echo "Starting documentation server..."
	@cd docs && python3 -m http.server 8000

# 验证文档
docs-validate:
	@echo "Validating OpenAPI specification..."
	spectral lint docs/swagger.yaml
	@echo "Validation passed"

# 检查文档覆盖率
docs-coverage:
	@echo "Checking documentation coverage..."
	go run tools/check-docs-coverage.go
```

### 使用方式

```bash
make docs           # 生成文档
make docs-serve     # 本地预览
make docs-validate  # 验证文档
make docs-coverage  # 检查覆盖率
```

## 4. 文档验证工具配置

### Spectral配置

**安装：**
```bash
npm install -g @stoplight/spectral-cli
```

**创建 `.spectral.yaml`：**
```yaml
extends: [[spectral:oas, all]]

rules:
  operation-summary: error
  operation-description: error
  operation-tags: error
  operation-operationId: error
  parameter-description: error
  response-description: error
  schema-example: warn

  # 自定义规则
  operation-success-response:
    description: "操作必须至少有一个2xx响应"
    severity: error
    given: "$.paths[*][*]"
    then:
      field: "responses"
      function: schema
      functionOptions:
        schema:
          type: object
          required: ["200", "201", "204"]
          minProperties: 1
```

**验证命令：**
```bash
spectral lint docs/swagger.yaml
```

## 5. CI/CD集成配置

### GitHub Actions配置

**创建 `.github/workflows/api-docs.yml`：**
```yaml
name: API Documentation

on:
  push:
    branches: [main]
    paths:
      - '**.go'
      - 'docs/**'
  pull_request:
    branches: [main]

jobs:
  generate-and-validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install swag
        run: go install github.com/swaggo/swag/cmd/swag@latest

      - name: Generate documentation
        run: swag init

      - name: Check for uncommitted changes
        run: |
          if [[ -n $(git status -s docs/) ]]; then
            echo "Documentation needs to be regenerated"
            echo "Run: swag init"
            exit 1
          fi

      - name: Install spectral
        run: npm install -g @stoplight/spectral-cli

      - name: Validate OpenAPI spec
        run: spectral lint docs/swagger.yaml --fail-severity=error

      - name: Check documentation coverage
        run: go run tools/check-docs-coverage.go

      - name: Upload docs artifact
        uses: actions/upload-artifact@v3
        with:
          name: api-docs
          path: docs/

  deploy:
    needs: generate-and-validate
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Download docs artifact
        uses: actions/download-artifact@v3
        with:
          name: api-docs
          path: docs/

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs
```

### GitLab CI配置

**创建 `.gitlab-ci.yml`：**
```yaml
stages:
  - generate
  - validate
  - deploy

generate-docs:
  stage: generate
  image: golang:1.21
  script:
    - go install github.com/swaggo/swag/cmd/swag@latest
    - swag init
  artifacts:
    paths:
      - docs/
    expire_in: 1 week

validate-docs:
  stage: validate
  image: node:18
  dependencies:
    - generate-docs
  script:
    - npm install -g @stoplight/spectral-cli
    - spectral lint docs/swagger.yaml --fail-severity=error

deploy-docs:
  stage: deploy
  image: alpine:latest
  dependencies:
    - generate-docs
  only:
    - main
  script:
    - apk add --no-cache rsync openssh
    - rsync -avz docs/ $DEPLOY_SERVER:/var/www/docs/
```

## 配置文件总结

### 项目结构

```
project/
├── .github/
│   └── workflows/
│       └── api-docs.yml
├── .spectral.yaml
├── .swaggo
├── docs/
│   ├── swagger.json
│   └── swagger.yaml
├── swagger-ui/
│   ├── index.html
│   └── config.js
├── tools/
│   └── check-docs-coverage.go
├── Makefile
└── main.go
```

### 快速开始

```bash
# 1. 安装工具
go install github.com/swaggo/swag/cmd/swag@latest
npm install -g @stoplight/spectral-cli

# 2. 生成文档
make docs

# 3. 验证文档
make docs-validate

# 4. 本地预览
make docs-serve

# 5. 检查覆盖率
make docs-coverage
```
