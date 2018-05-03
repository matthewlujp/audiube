.PHONY : install build run prepare-test test

install:
	go get github.com/golang/mock/gomock
	go get github.com/golang/mock/mockgen
	dep ensure

build:
	cd src && go build -o ../server .

run: | build
	./server --static ./static

prepare-test:
	@echo "\nPreparing test........................................"
	mockgen --source ./src/youtube_data_v3/client.go --destination ./src/youtube_data_v3/mocks/mock_client.go

test: | prepare-test
	@echo "\nRun test........................................"
	@cd src && go test 