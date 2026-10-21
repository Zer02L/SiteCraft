# Cookbook: Templ Component with Go CVA Variants

Canonical pattern for creating reusable UI primitives with zero-allocation type-safe variants and Tailwind v4 design tokens.

## File: `ui/components/base/<name>.templ`

```templ
package base

import "github.com/Oudwins/tailwind-merge-go"

type NoticeVariant uint8

const (
	NoticeInfo NoticeVariant = iota
	NoticeSuccess
	NoticeWarning
	NoticeDestructive
)

type NoticeProps struct {
	Variant NoticeVariant
	Class   string
	Attrs   templ.Attributes
}

func (p NoticeProps) Classes() string {
	const base = "flex items-start gap-3 p-4 rounded-card border text-sm transition-all"
	var variantClass string
	switch p.Variant {
	case NoticeSuccess:
		variantClass = "bg-emerald-500/10 border-emerald-500/20 text-emerald-600 dark:text-emerald-400"
	case NoticeWarning:
		variantClass = "bg-amber-500/10 border-amber-500/20 text-amber-600 dark:text-amber-400"
	case NoticeDestructive:
		variantClass = "bg-rose-500/10 border-rose-500/20 text-rose-600 dark:text-rose-400"
	default:
		variantClass = "bg-accent/10 border-accent/20 text-accent"
	}
	return twmerge.Merge(base, variantClass, p.Class)
}

templ Notice(props NoticeProps) {
	<div class={ props.Classes() } { props.Attrs... }>
		{ children... }
	</div>
}
```

## Important Rules:
1. Always run `templ generate` after editing any `.templ` file.
2. Use semantic tokens (`rounded-card`, `bg-accent`, `text-content`, etc.) rather than hardcoded pixel/hex values.
3. Accept optional `Class string` and `Attrs templ.Attributes` for composition and HTMX attributes.
