build: assets main.go
	go build -o filekeep .

run: assets main.go
	go run main.go

assets: build_assets.sh
	bash build_assets.sh
