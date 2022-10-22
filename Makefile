all: main.c wallet.c setup.py libwasmd.a libwasmd.h
	python3 setup.py build

libwasmd.a:
	./build-libwasmd.sh