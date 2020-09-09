# build
.PHONY : build txstorm
build :
	go build -o build/opera ./cmd/opera

txstorm :
	go build -o build/tx-storm ./cmd/tx-storm
#test
.PHONY : test
test :
	go test ./...

#clean
.PHONY : clean
clean :
	rm ./build/opera
