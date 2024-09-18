unit-test-pagination:
	@go test -v ./main.go ./itemcache.go ./mongo.go ./pagination.go ./pagination_test.go

unit-test-mongo:
	@go test -v ./main.go ./mongo.go ./mongo_test.go

unit-test-itemcache:
	@go test -v ./main.go ./itemcache.go ./itemcache_test.go