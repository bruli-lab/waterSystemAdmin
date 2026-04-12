module github.com/bruli/waterSystemAdmin

go 1.26.2

require (
	github.com/flosch/pongo2 v0.0.0-20200913210552-0d938eb266f3
	github.com/flosch/pongo2/v6 v6.0.0
	github.com/gorilla/sessions v1.4.0
	github.com/kelseyhightower/envconfig v1.4.0
	golang.org/x/crypto v0.50.0
	golang.org/x/net v0.53.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/telemetry v0.0.0-20240522233618-39ace7a40ae7 // indirect
	golang.org/x/tools v0.32.0 // indirect
	golang.org/x/vuln v1.1.4 // indirect
	mvdan.cc/gofumpt v0.8.0 // indirect
)

tool (
	golang.org/x/vuln/cmd/govulncheck
	mvdan.cc/gofumpt
)
