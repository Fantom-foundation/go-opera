current_dir = $(shell pwd)

clean:
	rm -rf *.pb.go

proto:
	protoc --go_out=plugins=grpc,import_path=poset:. *.proto
