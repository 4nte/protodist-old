module {{ .Module }}

go {{ .GoVersion }}

require (
	github.com/buger/jsonparser v0.0.0-20200322175846-f7e751efca13
	github.com/golang/protobuf v1.3.5
	github.com/paulmach/go.geojson v1.4.0
	github.com/pkg/errors v0.9.1
	google.golang.org/grpc v1.28.0
)