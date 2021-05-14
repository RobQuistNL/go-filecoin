export CGO_CFLAGS_ALLOW=-D__BLST_PORTABLE__
export CGO_CFLAGS=-D__BLST_PORTABLE__

all:
	go run ./build/*.go build

deps:
	git submodule update --init
	go run ./build/*.go smartdeps

# WARNING THIS BUILDS A GO PLUGIN AND PLUGINS *DO NOT* WORK ON WINDOWS SYSTEMS
iptb:
	make -C tools/iptb-plugins all

clean:
	rm ./venus

	rm -rf ./extern/filecoin-ffi
	rm -rf ./extern/test-vectors