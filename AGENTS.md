# AGENTS.md

This repository uses workpads for structured project context and progress tracking.
Cursor rules already define the workflow; this file mirrors those rules for agents.

## Workpads Overview

Workpads live under `workpads/` and each project has:

```
workpads/
  {project}/
    knowledge.md
    references.md
    tasks.md
    repos/
```

Progress is recorded in files and git, not chat context.

## Before Starting Any Work

1. Identify the active project in `workpads/README.md`.
2. Read these files for that project:
   - `workpads/{project}/knowledge.md`
   - `workpads/{project}/references.md`
   - `workpads/{project}/tasks.md`
3. If `workpads/{project}/repos/` is referenced but empty, follow setup in
   `workpads/{project}/references.md`.

## Task Protocol

Status markers:

```
📋 pending
🚧 in_progress
✅ completed
🚫 blocked
```

Before starting a task:
1. Mark the task as `🚧 in_progress`.
2. Read relevant sections of `knowledge.md`.
3. Consult `references.md` as needed.

During task execution:
- **Add discovered tasks** to `tasks.md` as you work.
- Prerequisites, follow-ups, investigations, or risks should become new tasks.
- Include clear descriptions and acceptance criteria for new tasks.

Before completing a task, verify:
- Acceptance criteria met.
- `knowledge.md` updated if decisions were made.
- `references.md` updated if new references were used.
- Any discovered tasks added to `tasks.md`.

After completing a task:
1. Mark the task `✅ completed`.
2. Update `knowledge.md` and `references.md` if needed.
3. Commit progress (unless the user says not to).
4. Request user confirmation before starting the next task.

## When to Update Docs

Update `knowledge.md` when:
- Design decisions, specs, breaking changes, or lessons learned are identified.

Update `references.md` when:
- New external resources are discovered or reference quality changes.

## Exception Handling

- Urgent fixes: may interrupt flow but must update `tasks.md`.
- Scope changes: update `tasks.md` before proceeding.
- Blocked tasks: mark as blocked and create unblocking tasks.
- Missing repos: clone into `workpads/{project}/repos/`.

## Source of Truth

See `.cursor/rules/workpads_workflow.mdc` for the canonical workflow.

## Cursor Commands

The `next` command follows this sequence (from `.cursor/commands/next.md`):

1. Read state files:
   - `workpads/README.md`
   - `workpads/{active-project}/tasks.md`
   - `workpads/{active-project}/knowledge.md`
   - `workpads/{active-project}/references.md`
2. Select a `📋 pending` task based on dependencies, state, risk, and testability.
3. Load context and ensure reference repos exist under
   `workpads/{active-project}/repos/` if needed.
4. Execute:
   - Mark task `🚧 in_progress`
   - Implement and test
   - Update `knowledge.md` and `references.md`
   - Mark task `✅ completed`
   - Commit when appropriate
   - Request confirmation before switching tasks
