.PHONY: all unigornel go
.PHONY: install

UNIGORNEL_YAML=${HOME}/.unigornel.yaml

all: unigornel go

unigornel:
	cd unigornel && go install
	@echo "[+] the unigornel binary is in ${GOPATH}/bin/unigornel"

go:
	cd go/src && GOOS=unigornel GOARCH=amd64 ./make.bash

install: unigornel
	rm -f $(UNIGORNEL_YAML)
	@echo "[+] writing $(UNIGORNEL_YAML)"
	@echo "goroot: ${PWD}/go" >> $(UNIGORNEL_YAML)
	@echo "minios: ${PWD}/minios" >> $(UNIGORNEL_YAML)
	@echo "libraries: ${PWD}/libraries.yaml" >> $(UNIGORNEL_YAML)
	@cat $(UNIGORNEL_YAML)
	@echo '[+] run `eval $$(unigornel env)` to setup the unigornel environment'

