module github.com/bhechinger/jack-exporter

go 1.15

require (
	github.com/prometheus/client_golang v1.9.0
	github.com/xthexder/go-jack v0.0.0-20201026211055-5b07fb071116
	go.uber.org/zap v1.13.0
)

replace github.com/xthexder/go-jack => ../go-jack
