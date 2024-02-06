
ORG := dmolik
BIN := pgweb-k8s-backend
GO := $(shell which go)
OPENSSL := $(shell which openssl)
SRC := $(shell find ./ -name \*.go -not -name \*_test.go)
VERSION ?= $(shell  if [ ! -z $$(git tag --points-at HEAD) ] ; then git tag --points-at HEAD|cat ; else  git rev-parse --short HEAD|cat; fi )
BUILD ?= $(shell git rev-parse --short HEAD)
REPO := github.com/$(ORG)/$(BIN)
IMAGE ?= $(REPO):$(VERSION)
RUNTIME ?= docker

V ?= 0
ifeq ($(V), 1)
	Q =
	VV = -v
else
	Q = @
	VV =
endif

build: $(BIN)

$(BIN): $(SRC) go.mod go.sum
	$Q$(GO) build $(VV) \
		-trimpath -asmflags all=-trimpath=/src -installsuffix cgo \
		-ldflags "-s -w -X $(REPO).Version=$(VERSION) -X $(REPO).Build=$(BUILD)" \
		-gcflags "all=-N -l" \
		-o $@ ./main/main.go

lib64:
	$Qinstall -d $@

libs: lib64 $(BIN)
	$(shell for i in $$(ldd $(BIN) | sed -e '/linux-vdso/d' -e 's/^[\ \t]\+//' -e 's/(0x[0-9a-h]\+)//' -e 's/^\S\+\ =>\ //' ) ; do cp $$i lib64 ; done)

.PHONY: container
container: libs $(BIN)
	$Q$(RUNTIME) build -t $(IMAGE) .

.PHONY: clean real-clean
clean:
	$Qrm -f $(BIN)

real-clean: clean
	$Q$(GO) clean -cache -testcache -modcache -i -r

.PHONY: check sec gosec trivy
check: lint sec
sec: trivy gosec

trivy: container
	$Qtrivy i --scanners vuln,misconfig,secret $(IMAGE)

gosec:
	$Qgosec -exclude-generated -exclude-dir mod -exclude-dir cache -exclude-dir tmp -exclude-dir go ./...

lint:
	$Qgolangci-lint run ./...

aes-key:
	$Qecho $(shell $(OPENSSL) enc -aes-256-cbc -k secret -P -md sha1  -pbkdf2 -iter 100000 | grep key | sed 's/key=//g' )
