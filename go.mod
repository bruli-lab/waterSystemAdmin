module github.com/bruli-lab/waterSystemAdmin

go 1.27.4

require (
	github.com/gorilla/sessions v1.4.0
	github.com/kelseyhightower/envconfig v1.4.0
	golang.org/x/crypto v0.59.0
)

require (
	github.com/gorilla/securecookie v1.1.2 // indirect
	golang.org/x/mod v0.43.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/telemetry v0.0.0-20261116191925-e77a43154acb // indirect
	golang.org/x/tools v0.53.0 // indirect
	golang.org/x/vuln v1.10.0 // indirect
	mvdan.cc/gofumpt v0.12.0 // indirect
)

tool (
	golang.org/x/vuln/cmd/govulncheck
	mvdan.cc/gofumpt
)
