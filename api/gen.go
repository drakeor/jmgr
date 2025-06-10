//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api.proto

package api

// Optional: keep tools recorded in go.mod; they no longer need to be
// referenced here now that the directive lives in a buildable file.
