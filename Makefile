.PHONY: all
all: opera

.PHONY: opera
opera:
	GIT_COMMIT=`git rev-list -1 HEAD` && \
	GIT_DATE=`git log -1 --date=short --pretty=format:%ct` && \
	go build \
	    -ldflags "-s -w -X github.com/Fantom-foundation/go-opera/cmd/opera/launcher.gitCommit=$${GIT_COMMIT} -X github.com/Fantom-foundation/go-opera/cmd/opera/launcher.gitDate=$${GIT_DATE}" \
	    -o build/opera \
	    ./cmd/opera

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm ./build/opera
