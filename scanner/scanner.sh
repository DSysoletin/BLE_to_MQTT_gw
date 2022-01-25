#!/bin/bash
#sleeploop="sleep 600 && killall -s 9 scanner"
(while true; do sleep 600; killall -s 9 scanner; done) &

while true
do
#$sleeploop &
#(sleep 600 && killall -s 9 scanner) &
#./scanner -host 192.168.0.2 -port 14419
/root/BLE_to_MQTT_gw/scanner/scanner -host 192.168.0.2 -port 14419
sleep 10
done
