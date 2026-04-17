PROTO_DIR := proto
GEN_DIR := gen
MODULE := dos 

PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto")

.PHONY: gen clean

gen:
	protoc -I $(PROTO_DIR) \
		--go_out=. \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(MODULE) \
		$(PROTO_FILES)

clean:
	rm -rf $(GEN_DIR)
