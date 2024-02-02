
ORG := dmolik
BIN := pgweb-k8s-backend
GO := $(shell which go)
OPENSSL := $(shell which openssl)
SRC := $(shell find ./ -name \*.go -not -name \*_test.go)
VERSION ?= $(shell  if [ ! -z $$(git tag --points-at HEAD) ] ; then git tag --points-at HEAD|cat ; else  git rev-parse --short HEAD|cat; fi )
BUILD ?= $(shell git rev-parse --short HEAD)
REPO := github.com/$(ORG)/$(BIN)

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

container: $(BIN)
	$Qdocker build -t $(REPO):$(VERSION) .

clean:
	$Qrm -f $(BIN)

real-clean: clean
	$Q$(GO) clean -cache -testcache -modcache -i -r

aes-key:
	$Qecho $(shell $(OPENSSL) enc -aes-256-cbc -k secret -P -md sha1  -pbkdf2 -iter 100000 | grep key | sed 's/key=//g' )
