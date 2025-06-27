run/new:
	go run ./main.go
run/latest:
	go run ./main.go -n=false
serve:
	go run ./main.go serve
list/models:
	go run ./main.go list
list/conversations:
	go run ./main.go conversation -l
build:
	go build -ldflags='-s' -o bin/clue main.go
