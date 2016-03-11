CPP = cpp
GOLANG = golang

PROTO_DIR = .
PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)
PROTO_GO = $(patsubst %.proto, %.pb.go, $(PROTO_FILES)) 
PROTO_HEADS = $(patsubst %.proto, %.pb.h, $(PROTO_FILES))
PROTO_CPPS = $(patsubst %.proto, %.pb.cc, $(PROTO_FILES))


all: $(GOLANG) $(CPP)
	@echo "compile all the proto files successfully"

$(GOLANG):
	mkdir -p $(GOLANG)
	for proto in $(PROTO_FILES); do\
		protoc -I $(PROTO_DIR) --go_out=$(PROTO_DIR)/$(GOLANG) $$proto;\
	done

$(CPP):
	mkdir -p $(CPP)
	@for proto in $(PROTO_FILES); do\
		protoc --proto_path=$(PROTO_DIR) --cpp_out=$(PROTO_DIR)/$(CPP) $$proto;\
	done


#%.pb.go: $(PROTO_DIR)/%.proto
#	protoc -I $(PROTO_DIR) --go_out=$(PROTO_DIR)/$(GOLANG) $<

#%.pb.cc: $(PROTO_DIR)/%.proto
#	protoc --proto_path=$(PROTO_DIR) --cpp_out=$(PROTO_DIR)/$(CPP) $<


clean:
	rm -rf $(PROTO_DIR)/$(GOLANG)
	rm -rf $(PROTO_DIR)/$(CPP)

show:
	@echo "=============== src ==========================="
	@echo PROTO_FILES: $(PROTO_FILES)
	@echo "=============== obj ==========================="
	@echo PROTO_GO: $(PROTO_GO)
	@echo PROTO_HEADS: $(PROTO_HEADS)
	@echo PROTO_CPPS: $(PROTO_CPPS)
