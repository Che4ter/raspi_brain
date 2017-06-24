package ipc

import (
	"fmt"
	"github.com/Che4ter/rpi_brain/configuration"
	"log"
	"net"
	"os"
)

//https://www.lucasamorim.ca/2016/03/20/unix-sockets-golang.html
func dataHandler(c net.Conn, ipcBridge chan string) {
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)

		if err != nil {
			return
		}

		data := string(buf[0 : nr])

		fmt.Println("received Data from unix socket:")

		fmt.Println(data)

		ipcBridge <- data
	}

	//ipcBridge <- data
}

func RunUnixSocketServer(ipcBridge chan string, config configuration.Configuration) {
	fmt.Println("starting Listen on Unix Socket:", config.UnixSocketAddress)

	if _, err := os.Stat(config.UnixSocketAddress); err == nil {
		// path/to/whatever exists
		err = os.Remove(config.UnixSocketAddress)
	}
	l, err := net.Listen("unix", config.UnixSocketAddress)
	if err != nil {
		log.Fatal(err)
		return
	}

	for {
		fd, err := l.Accept()
		if err != nil {
			log.Fatal(err)
			return
		}
		dataHandler(fd, ipcBridge)
		fd.Close()
	}
}
