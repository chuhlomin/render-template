The files are:
{{- $list := .added_files | split " " }}
{{- range $list }}
- {{ . }}
{{- end }}

Second list of files:
{{- $list = .added_files_2 | split "\n" }}
{{- range $list }}
- {{ . }}
{{- end }}

Third list of files:
{{- range .added_files_3 }}
- {{ . }}
{{- end }}
