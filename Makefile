.PHONY: run
run:
	go run *.go

.PHONY: curl-add
curl-add:
	 curl -i -X POST -d '{"value": 10 }' http://localhost:8080/add  -H "Content-Type: application/json"

.PHONY: curl-get
curl-get:
	 curl -i -X GET http://localhost:8080/get  -H "Content-Type: application/json"