#!/bin/sh
cd wasmd
sudo docker build -t libwasmd .
cd ..
id=$(sudo docker create libwasmd)
sudo docker cp $id:/code/libwasmd.a .
sudo docker cp $id:/code/libwasmd.h .
sudo docker rm -v $id