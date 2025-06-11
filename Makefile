

test:
	go test -v -cover ./...

buildGRPC:
	find ./internal -name "*.proto" -print0 | \
		xargs -0 -n1 \
		protoc -I . \
			--go_out=. \
			--go_opt=paths=source_relative \
			--go-grpc_out=. \
			--go-grpc_opt=paths=source_relative

