package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"
	"crypto/tls"
	"os"
	"os/signal"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/pkg/errors"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	//MQTT connection parameters
	usessl = flag.Bool("ssl", false, "Use encrypted MQTT connection if true")
	clientid = flag.String("clientid","golangIoTGW","MQTT Client ID")
	pempath = flag.String("pempath","./cert.pem","Path to cert.pem file")
	keypath = flag.String("keypath","./private.key","Path to private.key file")
	mqtthost = flag.String("host","myawsioturl.iot.us-west-2.amazonaws.com","MQTT server hostname")
	mqttport = flag.Int("port",8883,"MQTT server port")
	mqttpath = flag.String("mqttpath","/mqtt","MQTT server path")
	//BLE scanning params
	device = flag.String("device", "default", "implementation of ble")
	dup    = flag.Bool("dup", true, "allow duplicate reported")
)

type sensorData struct {
	macAddr string
	temperature float32
	humidity float32
	voltage int
	light int
	rssi int
}

var cdata = make(chan sensorData)
 
func mqttConnect(){
	cid := *clientid
	var connOpts MQTT.ClientOptions
	var brokerURL string

	if(*usessl){
		cer, err := tls.LoadX509KeyPair(*pempath, *keypath)
		check(err)
	
		connOpts = MQTT.ClientOptions{
		ClientID:             cid,
		CleanSession:         true,
		AutoReconnect:        true,
		MaxReconnectInterval: 1 * time.Second,
		KeepAlive:            int64(30 * time.Second),
		TLSConfig:            &tls.Config{Certificates: []tls.Certificate{cer}},
	}

	} else {
		connOpts = MQTT.ClientOptions{
			ClientID:             cid,
			CleanSession:         true,
			AutoReconnect:        true,
			MaxReconnectInterval: 1 * time.Second,
			KeepAlive:            int64(30 * time.Second),
		}
	}

	host := *mqtthost
	port := *mqttport
	path := *mqttpath
	if(*usessl){
		brokerURL = fmt.Sprintf("tcps://%s:%d%s", host, port, path)
	}else{
		brokerURL = fmt.Sprintf("tcp://%s:%d", host, port)
	}
	log.Println("[MQTT] debug: BrokerURL: %s",brokerURL)
	connOpts.AddBroker(brokerURL)

	mqttClient := MQTT.NewClient(&connOpts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Println("[MQTT] Connected")

	quit := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		mqttClient.Disconnect(250)
		fmt.Println("[MQTT] Disconnected")

		quit <- struct{}{}
	}()

	go func() {
		for{
		//var data sensorData
		data:= <- cdata
		fmt.Println("[MQTT] Publishing...")
		topic:=fmt.Sprintf("%s/temperature",data.macAddr)
		fmt.Println(topic)
		token := mqttClient.Publish(topic,0,false,fmt.Sprintf("%f",data.temperature))
		token.Wait();
		if token.Error() != nil {
                log.Fatal(token.Error())
        }
		topic=fmt.Sprintf("%s/humidity",data.macAddr)
		fmt.Println(topic)
		token = mqttClient.Publish(topic,0,false,fmt.Sprintf("%f",data.humidity))
		if token.Error() != nil {
                log.Fatal(token.Error())
        }

        topic=fmt.Sprintf("%s/voltage",data.macAddr)
        fmt.Println(topic)
		token = mqttClient.Publish(topic,0,false,fmt.Sprintf("%d",data.voltage))
		if token.Error() != nil {
                log.Fatal(token.Error())
        }

        topic=fmt.Sprintf("%s/light",data.macAddr)
        fmt.Println(topic)
		token = mqttClient.Publish(topic,0,false,fmt.Sprintf("%d",data.light))
		if token.Error() != nil {
                log.Fatal(token.Error())
        }

        topic=fmt.Sprintf("%s/rssi",data.macAddr)
        fmt.Println(topic)
		token = mqttClient.Publish(topic,0,false,fmt.Sprintf("%d",data.rssi))
		if token.Error() != nil {
                log.Fatal(token.Error())
        }

        }
	}()
	<-quit
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	go mqttConnect()

	d, err := dev.NewDevice(*device)
	if err != nil {
		log.Fatalf("can't use the BLE device : %s", err)
	}
	ble.SetDefaultDevice(d)
	//Start scanning
	fmt.Printf("Starting BLE scan...\n")
	ctx := ble.WithSigHandler(context.Background(),nil)
	chkErr(ble.Scan(ctx, *dup, advHandler, nil))
}

//Scan for advertisements
func advHandler(a ble.Advertisement) {
	var temperature float32
	var humidity float32
	var voltage int
	var int_part int
	var fract_part int
	var light int
	var sd ble.ServiceData
	var rssi int

	//fmt.Println("\n--------------------")
	//fmt.Printf("%+X\n", a)
	servData:=a.ServiceData()
	light=0
	if(len(servData)>0){
		sd=servData[0]
		//fmt.Printf("Service data length: %d \n",len(sd.Data))
		if (len(sd.Data) == 16) || (len(sd.Data) == 18) {
			fmt.Println("\n--------------------")
			int_part = int(sd.Data[4]) << 8 
			fract_part = int(sd.Data[5])
			temperature = float32(int_part + fract_part)/256
			if(temperature>200){
				temperature=temperature-256.0;
			}
			fmt.Printf("Temperature: %f\n", temperature)
			
			if(len(sd.Data)>14){
				int_part = int(sd.Data[14]) << 8 
				fract_part = int(sd.Data[15])
			}

			humidity = float32(int_part + fract_part)/256
			fmt.Printf("Humidity: %f\n", humidity)

			//voltage
			voltage = int(sd.Data[2]) << 8
			voltage += int(sd.Data[3])
			fmt.Printf("Voltage: %d\n", voltage)

			//light(optional)
			if (len(sd.Data) == 18) {
				light = int(sd.Data[16]) << 8
				light += int(sd.Data[17])
				fmt.Printf("Light: %d\n", light)
			}

			//RSSI
			rssi = a.RSSI()
			fmt.Printf("RSSI: %d\n", rssi)

			data:=sensorData{
				fmt.Sprintf("%s",a.Addr()),
				temperature,
				humidity,
				voltage,
				light,
				rssi,
			}
			cdata<-data
		}
	}


	//fmt.Println(a)
	//fmt.Printf("Underlying Type: %T\n", a)
    //fmt.Printf("Underlying Value: %v\n", a)


	/*if a.Connectable() {
		fmt.Printf("[%s] C %3d:", a.Addr(), a.RSSI())
	} else {
		fmt.Printf("[%s] N %3d:", a.Addr(), a.RSSI())
	}
	comma := ""
	if len(a.LocalName()) > 0 {
		fmt.Printf(" Name: %s", a.LocalName())
		comma = ","
	}
	if len(a.Services()) > 0 {
		fmt.Printf("%s Svcs: %v", comma, a.Services())
		comma = ","
	}
	if len(a.ManufacturerData()) > 0 {
		fmt.Printf("%s MD: %X", comma, a.ManufacturerData())
	}*/
	/*fmt.Printf(" Adv manufdata Len: %d ",len(a.ManufacturerData()))
	fmt.Printf(" Adv ServiceData Len: %d ",len(a.ServiceData()))
	fmt.Printf("\n")*/


}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		fmt.Printf("done\n")
	case context.Canceled:
		fmt.Printf("canceled\n")
	default:
		log.Fatalf(err.Error())
	}
}
