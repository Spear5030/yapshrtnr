module github.com/Spear5030/yapshrtnr

go 1.19

require (
	github.com/stretchr/testify v1.8.1
	internal/app v0.0.0-00010101000000-000000000000
	internal/handler v0.0.0-00010101000000-000000000000
	internal/storage v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace internal/app => ./internal/app

replace internal/storage => ./internal/storage

replace internal/handler => ./internal/handler
