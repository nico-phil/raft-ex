.PHONY: run-node-0
run-node-0:
	go run *.go \
	--node-id=node-0 \
	--raft-addr=127.0.0.1:7000 \
  	--http-addr=127.0.0.1:8000 \
  	--data-dir=node-0 \
  	--bootstrap=true

.PHONY: run-node-1
run-node-1:
	go run *.go \
	--node-id=node-1 \
	--raft-addr=127.0.0.1:7001 \
	--http-addr=127.0.0.1:8001 \
	--join-addr=127.0.0.1:8000 \
	--data-dir=node-1 \

.PHONY: run-node-2
run-node-2:
	go run *.go \
	--node-id=node-2 \
	--raft-addr=127.0.0.1:7002 \
	--http-addr=127.0.0.1:8002 \
	--join-addr=127.0.0.1:8000 \
	--data-dir=node-2 \

.PHONY: curl-set
curl-set:
	 curl -i -X POST -d '{"value": 10 }' http://127.0.0.1:8000/add  -H "Content-Type: application/json"

.PHONY: curl-get
curl-get:
	 curl -i -X GET http://127.0.0.1:8000/get  -H "Content-Type: application/json"

.PHONY: curl-join
curl-join:
	curl -i -X POST -d '{"node_id":" node-1", "raft_addr": "127.0.0.1:7000" }' http://localhost:8080/join  -H "Content-Type: application/json"

.PHONY: curl-getservers
curl-getservers:
	 curl -i -X GET http://127.0.0.1:8000/getservers  -H "Content-Type: application/json"

