# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hyperflow is an OpenSpec-based development environment. OpenSpec is an AI-native system for spec-driven development that uses a workflow of proposal → specs → design → tasks to structure software changes.

## Architecture

### Directory Structure

- `openspec/` - OpenSpec workspace root
  - `config.yaml` - Project configuration (schema: spec-driven)
  - `changes/` - Active change proposals being worked on
    - `archive/` - Completed changes (archived with date prefix: YYYY-MM-DD-<name>)
  - `specs/` - Main specification files for capabilities

### OpenSpec Workflow

The repository uses the "spec-driven" schema with these artifacts:
1. **proposal.md** - What and why (problem statement, goals, scope)
2. **specs/** - Detailed specifications for each capability
3. **design.md** - How (architecture, technical approach)
4. **tasks.md** - Implementation steps with checkboxes

### Change Lifecycle

Changes flow through these states:
- **Propose** - Create change with all artifacts (`/opsx:propose`)
- **Explore** - Think through problems and clarify requirements (`/opsx:explore`)
- **Apply** - Implement tasks from the change (`/opsx:apply`)
- **Archive** - Move completed change to archive (`/opsx:archive`)

## Commands

### OpenSpec CLI

The `openspec` CLI (v1.2.0) is available and required for all workflows.

**List changes:**
```bash
openspec list --json
```

**Check change status:**
```bash
openspec status --change "<name>" --json
```

**Get implementation instructions:**
```bash
openspec instructions apply --change "<name>" --json
```

**View available schemas:**
```bash
openspec schemas --json
```

### Workflow Commands

Use the Skill tool to invoke these workflows:

- `/opsx:propose` - Create a new change with all artifacts generated
- `/opsx:apply [change-name]` - Implement tasks from a change
- `/opsx:explore [change-name]` - Enter thinking mode to explore ideas
- `/opsx:archive [change-name]` - Archive a completed change

## Development Guidelines

### Working with Changes

1. **Always check status first** - Run `openspec status --change "<name>" --json` to understand the schema and artifact state
2. **Read context files** - Before implementing, read all files listed in `contextFiles` from `openspec instructions apply`
3. **Follow the schema** - The spec-driven schema defines artifact dependencies; respect them
4. **Mark tasks complete** - Update task checkboxes immediately: `- [ ]` → `- [x]`
5. **Keep changes minimal** - Implement only what's in the task, avoid over-engineering

### Artifact Creation

When creating artifacts (via `/opsx:propose` or manually):
- Use `openspec instructions <artifact-id> --change "<name>" --json` to get templates and rules
- The `template` field shows the structure to use
- `context` and `rules` are constraints for you - never copy them into the artifact file
- Read dependency artifacts before creating new ones

### Explore Mode

When in explore mode (`/opsx:explore`):
- **Never implement code** - This is for thinking, not coding
- Use ASCII diagrams liberally to visualize concepts
- May create OpenSpec artifacts (proposals, designs, specs) but not application code
- Ground discussions in the actual codebase when relevant

## Key Principles

- OpenSpec workflows are fluid, not phase-locked - you can update artifacts at any time
- Changes can be worked on incrementally - pause and resume as needed
- If implementation reveals design issues, update the design artifact
- Delta specs in changes can be synced to main specs during archive


## API设计规范
- 严格按照定义实现API
- 遵循 Microsoft REST API Guidelines（https://github.com/microsoft/api-guidelines/blob/vNext/azure/Guidelines.md）
- 每次修改增加或删除接口或者相关文件必须补全完整注释，不得影响swag文档阅读
- 成功响应直接返回资源对象，不使用 `{"data": ...}` 包装层
- 异步操作返回 202 + `Operation-Location` header，不直接暴露底层任务 ID
- 错误响应统一格式：`{"error": {"code": "PascalCase", "message": "..."}}`