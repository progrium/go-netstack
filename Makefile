
GVISOR_COMMIT=2691a8f9b1cf25d597905854569bae0a910c9964

gvisor:
	git clone --depth 1 https://github.com/google/gvisor
	cd gvisor && git fetch --depth 1 origin $(GVISOR_COMMIT)
	cd gvisor && git checkout $(GVISOR_COMMIT)
	rm -rf ./gvisor/.git