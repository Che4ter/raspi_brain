#!/bin/bash
echo cross compile for rpi3...
GOARM=7 GOARCH=arm GOOS=linux go build
echo finished build
scp rpi_brain pi@192.0.2.42:/home/pi/Pren2/logic
#scp rpi_brain pi@10.168.100.169:/home/pi/Pren2/logic

