package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
)

const END = '\x04'

type Connection struct {
	sock net.Conn
}

type VNResponse struct {
	Items []struct {
		Aliases     string   `json:"aliases"`
		Description string   `json:"description"`
		ID          int      `json:"id"`
		Image       string   `json:"image"`
		ImageNsfw   bool     `json:"image_nsfw"`
		Languages   []string `json:"languages"`
		Length      int      `json:"length"`
		Links       struct {
			Encubed   string      `json:"encubed"`
			Renai     interface{} `json:"renai"`
			Wikipedia string      `json:"wikipedia"`
		} `json:"links"`
		OrigLang  []string    `json:"orig_lang"`
		Original  interface{} `json:"original"`
		Platforms []string    `json:"platforms"`
		Released  string      `json:"released"`
		Title     string      `json:"title"`
	} `json:"items"`
	More bool `json:"more"`
	Num  int  `json:"num"`
}

func trim(r []byte) []byte {
	return bytes.TrimRight(bytes.TrimPrefix(r, []byte("results ")), string(END))
}

// OpenPlain opens a plain TCP connection to the API.
// Do not use this if you are transmitting your credentials.
func OpenPlain() (*Connection, error) {
	conn := new(Connection)
	s, err := net.Dial("tcp", "api.vndb.org:19534")
	if err != nil {
		return conn, errors.New("OpenPlain(): error opening connection: %s" + err.Error())
	}
	conn.sock = s
	return conn, nil
}

// Open opens a TLS connection to the API. This should be used as the default.
func Open() (*Connection, error) {
	conn := new(Connection)
	s, err := tls.Dial("tcp", "api.vndb.org:19535", &tls.Config{})
	if err != nil {
		return conn, errors.New("Open(): error opening connection: %s" + err.Error())
	}
	conn.sock = s
	return conn, nil
}

// Login logs the current connection into the given credentials.
// Returns nil if login was "ok"
func (c Connection) Login(username, password string) error {
	if len(username) == 0 || len(password) == 0 {
		return errors.New("Login(): zero length credential supplied")
	}
	fmt.Fprintf(c.sock, `login {"protocol":1,"client":"go-vndb","clientver":0.1,"username":"%s","password":"%s"}`+string(END), username, password)
	reply, err := bufio.NewReader(c.sock).ReadString(END)
	if err != nil {
		return errors.New("Login(): error reading reply: %s" + err.Error())
	}
	if reply != "ok"+string(END) {
		return errors.New("Login(): error logging in: %s" + reply)
	}
	return nil
}

// DBstats returns JSON VNDB statistics.
func (c Connection) DBStats() string {
	fmt.Fprint(c.sock, "dbstats"+string(END))
	reply, err := bufio.NewReader(c.sock).ReadString(END)
	if err != nil {
		fmt.Println("Error reading")
	}
	return reply
}

// GetVN gets a VN by its ID. Should return a VN object and nil.
func (c Connection) GetVN(id string) (*VNResponse, error) {
	vnr := new(VNResponse)
	if id == "" {
		return vnr, errors.New("GetVN(): ID cannot be empty.")
	}
	fmt.Fprintf(c.sock, "get vn basic,details (id = "+id+")"+string(END))
	reply, err := bufio.NewReader(c.sock).ReadSlice(END)
	if err != nil {
		return vnr, errors.New("GetVN(): Error reading string from connection: " + err.Error())
	}

	err = json.Unmarshal(trim(reply), vnr)
	if err != nil {
		return vnr, errors.New("GetVN(): Error unmarshalling JSON: " + err.Error())
	}

	return vnr, nil

}

func main() {
	user := flag.String("u", "foo", "username")
	pass := flag.String("p", "foo", "password")
	flag.Parse()

	conn, err := Open()
	defer conn.sock.Close()
	if err != nil {
		fmt.Println(err)
	}
	err = conn.Login(*user, *pass)
	if err != nil {
		fmt.Println(err)
	}

	vnr, err := conn.GetVN("11")
	if err != nil {
		panic(err)
	}
	fmt.Println(vnr.Items[0].Title)
}
