module github.com/things-kit/module/grpc

go 1.21

replace github.com/things-kit/module/log => ../log

require (
	github.com/spf13/viper v1.21.0
	github.com/things-kit/module/log v0.0.0-00010101000000-000000000000
	go.uber.org/fx v1.24.0
	google.golang.org/grpc v1.60.1
)

require (
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.16.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231002182017-d307bd883b97 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
