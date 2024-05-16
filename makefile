run:
	@rm -f *_templ.go
	@templ generate
	@go run .
