# API审查流程

## API设计审查清单

### 资源命名
- [ ] 使用复数名词（users, orders）
- [ ] 使用kebab-case命名（user-profiles）
- [ ] 路径中不包含动词
- [ ] 嵌套层级不超过3层
- [ ] 包含版本号（/api/v1/）

### HTTP方法
- [ ] GET用于获取资源
- [ ] POST用于创建资源
- [ ] PUT用于完整更新
- [ ] PATCH用于部分更新
- [ ] DELETE用于删除资源
- [ ] 方法语义正确

### HTTP状态码
- [ ] 200用于成功获取/更新
- [ ] 201用于成功创建（包含Location头）
- [ ] 204用于成功删除
- [ ] 400用于参数错误
- [ ] 401用于未认证
- [ ] 403用于无权限
- [ ] 404用于资源不存在
- [ ] 409用于资源冲突
- [ ] 429用于限流
- [ ] 500用于服务器错误

### 版本控制
- [ ] 使用URI路径版本（/api/v1/）
- [ ] 版本号从v1开始
- [ ] 破坏性变更增加主版本号

### 请求和响应
- [ ] 使用JSON格式
- [ ] 字段使用camelCase命名
- [ ] 日期时间使用ISO 8601格式
- [ ] 布尔值使用true/false
- [ ] 空值使用null

---

## API实现审查清单

### 错误处理
- [ ] 使用统一的错误响应格式
- [ ] 错误代码使用UPPER_SNAKE_CASE
- [ ] 错误消息清晰易懂
- [ ] 字段验证错误包含details数组
- [ ] 不暴露敏感信息（堆栈跟踪）

### 分页
- [ ] 支持offset和limit参数
- [ ] offset默认0，limit默认20
- [ ] limit最大值100
- [ ] 响应包含total、offset、limit
- [ ] 空结果返回空数组，不返回404

### 过滤和排序
- [ ] 支持常用字段过滤
- [ ] 支持sort参数（-前缀表示降序）
- [ ] 过滤参数验证
- [ ] 为过滤字段创建索引

### 认证
- [ ] 使用Bearer Token认证
- [ ] Access Token有效期15分钟
- [ ] Refresh Token有效期7天
- [ ] 支持Token刷新
- [ ] 未认证返回401

### 授权
- [ ] 实现基于角色的访问控制
- [ ] 验证资源所有权
- [ ] 无权限返回403
- [ ] 权限检查在服务端

### 限流
- [ ] 实现请求限流
- [ ] 返回限流响应头（X-RateLimit-*）
- [ ] 超限返回429和Retry-After
- [ ] 不同用户类型不同限流

### 幂等性
- [ ] PUT/DELETE操作幂等
- [ ] PATCH操作幂等
- [ ] POST支持Idempotency-Key（可选）

### 安全
- [ ] 使用HTTPS
- [ ] 验证输入参数
- [ ] 防止SQL注入
- [ ] 防止XSS攻击
- [ ] 敏感字段不返回（password）
- [ ] 实现CORS配置

---

## API文档审查清单

### OpenAPI规范
- [ ] 使用OpenAPI 3.0+
- [ ] 包含info对象（title, version, description）
- [ ] 定义servers
- [ ] 定义tags
- [ ] 定义securitySchemes

### 端点文档
- [ ] 所有端点都有summary
- [ ] 所有端点都有description
- [ ] 所有端点都有operationId
- [ ] 所有端点都有tags
- [ ] 所有参数都有描述
- [ ] 所有请求体都有schema
- [ ] 所有响应都有描述

### 数据模型
- [ ] 所有schema都有描述
- [ ] 所有属性都有类型
- [ ] 所有属性都有example
- [ ] 必需字段标记为required
- [ ] 枚举值已定义

### 示例
- [ ] 提供请求示例
- [ ] 提供响应示例
- [ ] 提供错误响应示例
- [ ] 示例数据真实可用

### 完整性
- [ ] 所有端点都已文档化
- [ ] 所有错误响应都已文档化
- [ ] 文档覆盖率100%
- [ ] Spectral验证通过

---

## API审查流程文档

### 审查时机

#### 1. 设计阶段审查
**时机：** 编写代码前

**目的：** 确保API设计符合规范

**参与者：** API设计者、技术负责人

**流程：**
1. 提交API设计文档（OpenAPI规范）
2. 技术负责人审查设计
3. 使用设计审查清单检查
4. 提出修改建议
5. 设计者修改后重新提交
6. 审查通过后开始实现

#### 2. 实现阶段审查
**时机：** 代码实现完成后

**目的：** 确保实现符合设计和规范

**参与者：** 开发者、Code Reviewer

**流程：**
1. 提交Pull Request
2. 自动运行CI检查（文档生成、验证）
3. Code Reviewer审查代码
4. 使用实现审查清单检查
5. 测试API功能
6. 提出修改建议
7. 开发者修改后重新提交
8. 审查通过后合并

#### 3. 文档审查
**时机：** API文档更新后

**目的：** 确保文档完整准确

**参与者：** 文档维护者、技术写作

**流程：**
1. 生成API文档
2. 检查文档覆盖率
3. 使用文档审查清单检查
4. 验证示例代码可运行
5. 检查文档可访问性
6. 提出修改建议
7. 修改后重新审查
8. 审查通过后发布

### 审查人员

- **技术负责人**：审查API设计
- **Code Reviewer**：审查代码实现
- **安全专家**：审查安全相关API
- **文档维护者**：审查API文档

### 审查标准

#### 必须通过（Blocker）
- 违反RESTful规范
- 存在安全漏洞
- 破坏向后兼容性
- 文档缺失或错误

#### 建议修改（Major）
- 命名不规范
- 错误处理不完善
- 性能问题
- 文档不完整

#### 可选优化（Minor）
- 代码风格
- 注释不足
- 可读性问题

### 审查工具

- **自动化检查**：CI/CD、Spectral、文档覆盖率工具
- **手动审查**：审查清单、Code Review
- **测试工具**：Postman、Swagger UI

---

## 审查反馈模板

### Pull Request审查反馈

```markdown
## API审查反馈

### 设计问题

#### Blocker
- [ ] 资源路径使用了单数名词 `/user` 应改为 `/users`
- [ ] 缺少版本号，应使用 `/api/v1/users`

#### Major
- [ ] 建议添加分页支持（offset/limit参数）
- [ ] 错误响应格式不统一

#### Minor
- [ ] 建议为status参数添加枚举值说明

### 实现问题

#### Blocker
- [ ] 未验证用户权限，存在安全风险
- [ ] 密码字段在响应中暴露

#### Major
- [ ] 缺少请求限流
- [ ] 未实现幂等性保证

#### Minor
- [ ] 建议添加日志记录

### 文档问题

#### Blocker
- [ ] 缺少Swagger注解，无法生成文档
- [ ] 错误响应未文档化

#### Major
- [ ] 缺少请求示例
- [ ] 参数描述不清晰

#### Minor
- [ ] 建议添加更多使用示例

### 总体评价

- 文档覆盖率: 85% (目标: 100%)
- Spectral验证: 3个错误，5个警告
- 建议: 修复所有Blocker问题后重新提交

### 下一步

1. 修复所有Blocker问题
2. 运行 `swag init` 重新生成文档
3. 运行 `spectral lint` 验证文档
4. 重新提交审查
```

### 设计审查反馈

```markdown
## API设计审查反馈

**API名称:** 用户管理API v1
**审查人:** 张三
**审查日期:** 2024-03-19

### 审查结果

- [ ] 通过
- [x] 需要修改
- [ ] 拒绝

### 问题列表

1. **资源命名** (Blocker)
   - 问题: 使用了动词 `/createUser`
   - 建议: 改为 `POST /users`

2. **版本控制** (Blocker)
   - 问题: 缺少版本号
   - 建议: 使用 `/api/v1/users`

3. **分页** (Major)
   - 问题: 未实现分页
   - 建议: 添加offset/limit参数

4. **错误处理** (Major)
   - 问题: 错误响应格式不统一
   - 建议: 使用标准错误格式

### 优点

- 资源模型设计合理
- 认证方案清晰
- 文档较完整

### 建议

1. 参考API设计标准文档修改
2. 添加完整的错误响应定义
3. 实现分页、过滤、排序功能

### 下一步

请在3个工作日内修改并重新提交审查。
```

---

## 审查流程自动化

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running API documentation checks..."

# 生成文档
swag init

# 检查文档变更
if [[ -n $(git status -s docs/) ]]; then
    git add docs/
    echo "✓ API documentation updated"
fi

# 验证文档
spectral lint docs/swagger.yaml --fail-severity=error
if [ $? -ne 0 ]; then
    echo "✗ OpenAPI validation failed"
    exit 1
fi

echo "✓ All checks passed"
```

### GitHub Actions审查

```yaml
name: API Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Generate docs
        run: swag init

      - name: Check coverage
        id: coverage
        run: |
          COVERAGE=$(go run tools/check-docs-coverage.go | grep "覆盖率" | awk '{print $2}')
          echo "coverage=$COVERAGE" >> $GITHUB_OUTPUT

      - name: Validate spec
        run: spectral lint docs/swagger.yaml

      - name: Comment PR
        uses: actions/github-script@v6
        with:
          script: |
            const coverage = '${{ steps.coverage.outputs.coverage }}';
            const body = `## 🔍 API审查报告\n\n- 文档覆盖率: ${coverage}\n- Spectral验证: 通过\n\n请确保所有审查清单项都已完成。`;
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body
            });
```
