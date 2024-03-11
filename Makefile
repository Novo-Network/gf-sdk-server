
define set_cgo_flags
export CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"
export CGO_CFLAGS="-O -D__BLST_PORTABLE__"
endef

build:
	go build -o gf-sdk-server src/main.go

testnet-run:
	$(call set_cgo_flags)
	./gf-sdk-server -private_key_path=$(HOME)/.ssh/gf-sdk-server.pk

mainnet-run:
	$(call set_cgo_flags)
	./gf-sdk-server -private_key_path=$(HOME)/.ssh/gf-sdk-server.pk -chain_rpc="" -chain_id=""

