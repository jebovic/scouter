package export

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

var mdTmpl = template.Must(template.New("mission").Funcs(template.FuncMap{
	"formatTime": func(t time.Time) string {
		return t.UTC().Format("2006-01-02")
	},
	"currency": func(amount float64, cur string) string {
		return fmt.Sprintf("%.2f %s", amount, cur)
	},
}).Parse(`# {{ .Mission.Name }}

**Category**: {{ .Mission.Category }}
**Budget**: {{ currency .Mission.Budget .Mission.Currency }}
**Phase**: {{ .Mission.Phase }}
**Exported**: {{ formatTime .ExportedAt }}

---

## Constraints

{{ if .Mission.Constraints }}
| Key | Label | Type | Value |
|-----|-------|------|-------|
{{ range .Mission.Constraints -}}
| {{ .Key }} | {{ .Label }} | {{ .Type }} | {{ .Value }} |
{{ end }}
{{ else }}
_No constraints defined._
{{ end }}

---

## Options ({{ len .Options }})

{{ range .Options }}
### {{ .Name }} — {{ .Badge }}

{{ if .PriceRange }}
**Price**: {{ currency .PriceRange.Best $.Mission.Currency }} ({{ currency .PriceRange.Min $.Mission.Currency }} – {{ currency .PriceRange.Max $.Mission.Currency }})
{{ end }}
{{ if .Notes }}{{ .Notes }}{{ end }}
{{ if .URL }}[View product]({{ .URL }}){{ end }}

{{ if .Attributes }}
| Attribute | Value |
|-----------|-------|
{{ range .Attributes -}}
| {{ .Label }} | {{ .Value }} |
{{ end }}
{{ end }}
{{ end }}

---

## Shopping List ({{ len .ShoppingItems }})

| Item | Merchant | Category | Price | Status |
|------|----------|----------|-------|--------|
{{ range .ShoppingItems -}}
| {{ .Name }} | {{ .Merchant }} | {{ .CostCategory }} | {{ currency .Price $.Mission.Currency }} | {{ .Status }} |
{{ end }}

---

## Decision Summary

{{ if .Decision }}
**Generated**: {{ formatTime .Decision.CreatedAt }}

{{ .Decision.Summary }}

### Scores

| Option | Score | Price | Quality | Features | Eliminated |
|--------|-------|-------|---------|----------|------------|
{{ range .Decision.Scores -}}
| {{ .OptionName }} | {{ printf "%.1f" .Score }} | {{ printf "%.1f" .PriceScore }} | {{ printf "%.1f" .QualScore }} | {{ printf "%.1f" .FeatScore }} | {{ if .Eliminated }}⚠ {{ .EliminReason }}{{ else }}No{{ end }} |
{{ end }}
{{ else }}
_No decision has been run for this mission._
{{ end }}

---

_Exported from SCOUTER on {{ formatTime .ExportedAt }}_
`))

// ToMarkdown renders ExportData as a Markdown document.
func ToMarkdown(data *ExportData) ([]byte, error) {
	var buf bytes.Buffer
	if err := mdTmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("export markdown: %w", err)
	}
	return buf.Bytes(), nil
}
