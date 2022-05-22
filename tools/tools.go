//go:build tools

package tools

import (
	_ "github.com/daixiang0/gci"
	_ "github.com/golang/mock/mockgen"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
	_ "mvdan.cc/gofumpt"
)
