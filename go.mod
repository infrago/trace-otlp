module github.com/infrago/trace-otlp

go 1.25.3

require (
	github.com/infrago/base v0.10.0
	github.com/infrago/infra v0.10.0
	github.com/infrago/trace v0.10.0
)

require github.com/pelletier/go-toml/v2 v2.2.2 // indirect

replace github.com/infrago/infra => ../bamgoo

replace github.com/infrago/base => ../base

replace github.com/infrago/trace => ../trace
