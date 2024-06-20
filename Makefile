
GVISOR_COMMIT=2691a8f9b1cf25d597905854569bae0a910c9964

gvisor:
	git clone --depth 1 https://github.com/google/gvisor
	cd gvisor && git fetch --depth 1 origin $(GVISOR_COMMIT)
	cd gvisor && git checkout $(GVISOR_COMMIT)
	rm -rf ./gvisor/.git

UIO_COMMIT=82958018845cfb6b02c09e57c390e67c1d81ee95

uio:
	git clone --depth 1 https://github.com/u-root/uio
	cd uio && git fetch --depth 1 origin $(UIO_COMMIT)
	cd uio && git checkout $(UIO_COMMIT)
	rm -rf ./uio/.git

DHCP_COMMIT=1ca156eafb9f20f7884eddc2cf610bade5dfb560

dhcp:
	git clone --depth 1 https://github.com/insomniacslk/dhcp
	cd dhcp && git fetch --depth 1 origin $(DHCP_COMMIT)
	cd dhcp && git checkout $(DHCP_COMMIT)
	rm -rf ./dhcp/.git