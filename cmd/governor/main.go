package main

import (
	"net"
	"net/http"
	"os"

	"github.com/ParadropLabs/governor/pkg/governor"
)

const (
	DefaultSocket = "/var/run/governor.socket"
	DefaultSnapdSocket = "/var/run/snapd.socket"
)

func main() {
	var socket string
	var snapdSocket string
	var ok bool

	socket, ok = os.LookupEnv("GOVERNOR_SOCKET")
	if !ok {
		socket = DefaultSocket
	}

	snapdSocket, ok = os.LookupEnv("SNAPD_SOCKET")
	if !ok {
		snapdSocket = DefaultSnapdSocket
	}

	snapdProxy := governor.NewSnapdProxy("/snapd", snapdSocket)
	usersResource := governor.NewUsersResource("/users")

	http.HandleFunc("/snapd/", snapdProxy.ServeHTTP)
	http.HandleFunc("/users/", usersResource.ServeHTTP)

	os.Remove(socket)

	listener, err := net.Listen("unix", socket)
	if err != nil {
		panic(err)
	}

	http.Serve(listener, nil)
}
