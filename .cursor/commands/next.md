# Next: Do Next Task

Follow the workpads workflow and complete the next task.

## Step 1: Read State Files

Read these first:
1. `workpads/README.md` - Active projects
2. `workpads/{active-project}/tasks.md` - Current tasks
3. `workpads/{active-project}/knowledge.md` - Decisions and context
4. `workpads/{active-project}/references.md` - External references and notes

## Step 2: Select a Task

Choose a `📋 pending` task based on:
- Dependencies
- Current state
- Risk and unknowns
- Testability

## Step 3: Load Context

Follow any project-specific instructions in `knowledge.md` and `references.md`.
If reference repos are required, ensure they are placed under
`workpads/{active-project}/repos/` (gitignored).
Keep `tasks.md` in git to preserve progress history.

## Step 4: Execute

1. Mark task as `🚧 in_progress` in `tasks.md`
2. Implement the task and run relevant tests
3. Update `knowledge.md` with decisions or learnings
4. Update `references.md` with new references and quality notes
5. Mark task as `✅ completed` in `tasks.md`
6. Commit progress to git when appropriate
7. Request confirmation before starting a different task

Start now.
