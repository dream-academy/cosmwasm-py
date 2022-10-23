all: main.c wallet.c setup.py libwasmd.a libwasmd.h
	python3 setup.py install

libwasmd.a:
	LEDGER_ENABLED=false LINK_STATICALLY=true make -C wasmd libwasmd
	cp wasmd/libwasmd.a .
	cp wasmd/libwasmd.h .

clean:
	rm -rf build libwasmd.a libwasmd.h