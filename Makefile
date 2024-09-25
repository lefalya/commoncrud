unit-test-pagination:
	@go test -v ./main.go ./itemcache.go ./mongo.go ./pagination.go ./pagination_test.go

unit-test-mongo:
	@go test -v ./main.go ./mongo.go ./mongo_test.go

unit-test-itemcache:
	@go test -v ./main.go ./itemcache.go ./itemcache_test.go

integration-test:
	@go test -v ./main.go ./itemcache.go ./mongo.go ./pagination.go ./pagination_integration_test.go

test-coverage:
	@go test -v ./main.go ./itemcache.go ./mongo.go ./pagination.go ./itemcache_test.go ./mongo_test.go ./pagination_test.go -coverprofile=coverage.out
	@go tool cover -html=coverage.out

mock-interfaces:
	@mockgen -source=interfaces/main.go --destination=./mocks/interfaces.go