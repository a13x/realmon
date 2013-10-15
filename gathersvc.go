package main

import (
	"code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"flag"
	"fmt"
	eventsource "github.com/antage/eventsource/http"
	"log"
	"net"
	"net/http"
	"spotify.com/realmon"
	"time"
)

type PhoneIn struct {
	Addr     *net.UDPAddr
	Report   *spotify_realmon.Report
	Reported time.Time
}

func (p PhoneIn) String() string {
	return fmt.Sprintf("(%s %s)", p.Addr, p.Report)
}

// shorthand helper functions
func (p PhoneIn) Service() string {
	return p.Report.GetService()
}

func (p PhoneIn) Id() string {
	return string(p.Report.Uuid[:])
}

func (p PhoneIn) StatusUrl() string {
	return p.Report.GetStatusUrl()
}

type Services map[string]map[string]*SrvInstance

func (s Services) update(service, instance string, val *SrvInstance) {
	_, exists := s[service]
	if !exists {
		ms := make(map[string]*SrvInstance)
		s[service] = ms
	}
	s[service][instance] = val
}

type SrvInstance struct {
	Id           string
	Status_url   string
	Last_phonein PhoneIn
	Age          int64
}

func read(conn *net.UDPConn, ch chan PhoneIn) {
	for {
		b := make([]byte, 512)
		bread, addr, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Fatal("aaaargh!")
		}
		report := &spotify_realmon.Report{}
		err = proto.Unmarshal(b[:bread], report)
		if err != nil {
			log.Fatal("couldn't do it, sorry!")
		}
		phonein := PhoneIn{addr, report, time.Now()}
		ch <- phonein
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	rpc <- "status"
	result := <-rpc
	fmt.Fprintf(w, "%s", result)
}

func UdpReceiver(addr string, rpc chan string) {
	maddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatal("can't resolve UDP address ", err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, maddr)
	if err != nil {
		log.Fatal("can't listen for udp packet ", err, addr)
	}
	services := Services{}
	ch := make(chan PhoneIn)
	go read(conn, ch)
	defer conn.Close()
	for {
		select {
		case in := <-ch:
			instance := &SrvInstance{
				Id:           in.Id(),
				Status_url:   in.StatusUrl(),
				Last_phonein: in}
			services.update(in.Service(), in.Id(), instance)
		case <-rpc:
			b, err := json.Marshal(services)
			if err != nil {
				fmt.Println("ups!")
			}
			rpc <- string(b[:])
		}
	}
}

var rpc = make(chan string)

func main() {
	var port = flag.Int("p", 9000, "UDP Port")
	flag.Parse()
	laddr := fmt.Sprintf("239.255.13.0:%d", *port)
	go UdpReceiver(laddr, rpc)
	es := eventsource.New(nil)
	defer es.Close()
	http.Handle("/", http.FileServer(http.Dir("./pub")))
	http.HandleFunc("/ph", statusHandler)
	http.Handle("/events", es)
	go func() {
		for {
			rpc <- "events"
			result := <-rpc
			es.SendMessage(result, "realmon", "")
			time.Sleep(time.Second)
		}
	}()
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Cannot serve web")
	}
}
