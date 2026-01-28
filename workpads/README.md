# Workpads

Structured project documentation for complex tasks requiring research, planning, and execution.

## Active Projects

| Project | Path | Status | Description |
|---------|------|--------|-------------|
| **Gaia Migration** | `workpads/gaia-migration/` | 🚧 Active | Migrate Osmosis DEX modules to Gaia |

---

## Structure

Each project folder contains:

```
workpads/
├── {project}/
│   ├── knowledge.md    # Internal decisions, specs, architecture
│   ├── references.md   # External links, research, patterns  
│   ├── tasks.md        # Current task list (committed)
│   └── repos/          # Reference repositories (gitignored, optional)
└── README.md
```

## Standard Files

| File | Purpose | Committed? |
|------|---------|------------|
| `knowledge.md` | Internal decisions, architecture, lessons learned | ✅ Yes |
| `references.md` | Curated external references with quality notes | ✅ Yes |
| `tasks.md` | Current task list with acceptance criteria | ✅ Yes |
| `repos/` | Reference repository clones for comparison | ❌ No (gitignored) |

## Workflow

See `.cursor/rules/workpads_workflow.mdc` for the complete workflow guide.
