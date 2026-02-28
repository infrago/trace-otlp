module github.com/bamgoo/trace-otlp

go 1.25.3

require (
	github.com/bamgoo/bamgoo v0.0.1
	github.com/bamgoo/base v0.0.1
	github.com/bamgoo/trace v0.0.1
)

require github.com/pelletier/go-toml/v2 v2.2.2 // indirect

replace github.com/bamgoo/bamgoo => ../bamgoo

replace github.com/bamgoo/base => ../base

replace github.com/bamgoo/trace => ../trace
