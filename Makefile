all:
	go build -ldflags="-w -s" -o annotationTool main.go

clean:
	rm -f annotationTool