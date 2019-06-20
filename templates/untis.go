package templates

// copied and modified from reddec/system-gen project
import "text/template"

var ServiceUnitTemplate = template.Must(template.New("").Parse(`[Unit]
Description={{.Name}}

[Service]
{{- range $k, $v :=  .Environment}}
Environment={{$k}}={{$v}}
{{- end}}
ExecStart={{.Command}}
Restart=always
RestartSec=5
{{- with .WorkingDirectory}}
WorkingDirectory={{.}}
{{- end}}

[Install]
WantedBy=multi-user.target
`))
