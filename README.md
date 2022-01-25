# BLE_to_MQTT_gw
Simple BLE to MQTT gateway

This app will scan for the Eddystone BLE advertisement packets from my ESP32 BLE sensor, and forward temperature, voltage and humidity values to the cloud.

How to use:
1) Please upload my ESP32 firmware (https://github.com/DSysoletin/ESP32_BLE_env_sensor) to your ESP32 board and connect the DHT11 sensor 

2) If you want to use the binary, just skip this step. If you want to built the scanner, please do the following commands (from the clonned repo):
$cd ./scanner
$go build

3) To use the scanner with local MQTT server, you just need to specify the host and port (please note that scanner needs to be run from root user because it needs to have access to the hci device directly, not via BlueZ):
#./scanner -host 127.0.0.1 -port 1883

4) To use the scanner with AWS IoT Core, please supply your certificates:
#./scanner -host YOUR_ENDPOINT_URL.amazonaws.com -ssl -keypath ./IoT-sensor-test.private.key -pempath ./IoT-sensor-test.cert.pem

When the scanner will detect some Eddystone messages and parse it, it will publish them to the MQTT server. You'll see prints like this:
Publishing...
30:83:98:00:3c:2a/temperature
30:83:98:00:3c:2a/humidity
30:83:98:00:3c:2a/voltage

In this example, you can see the topics used for publishing the sensor data. Because topics include HW address of the sensor, several sensors can be used with one scanner app.

AUTOSTART (Gentoo only)

I've created a script for autostart the BLE scanner via OpenRC.
To use it, just copy "gentoo_init_script/ble_scanner" to "/etc/init.d" directory.
Please make sure that you clonned repo in your root folder (ls /root/BLE_to_MQTT_gw/scanner/scanner should show valid application), OR edit the ble_scanner and scanner.sh scritps and replace paths to your actual ones.
And please make sure that scanner binary, scanner.sh and ble_scanner files have +x flag

After that, you can start and stop the scanner with /etc/init.d/ble_scanner start(or stop)
