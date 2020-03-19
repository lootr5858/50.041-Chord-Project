package chord

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
)

// Between checks if identifier is in range (a, b)
func Between(nodeX, nodeA, nodeB int) bool {
	if nodeB < nodeA && nodeX > nodeB {
		nodeB += ringSize
	} else if nodeB < nodeA && nodeX < nodeB {
		nodeA -= ringSize
	}
	return nodeX > nodeA && nodeX < nodeB
}

// BetweenRightIncl checks if identifier is in range (a, b]
func BetweenRightIncl(nodeX, nodeA, nodeB int) bool {
	if nodeB < nodeA && nodeX > nodeB {
		nodeB += ringSize
	} else if nodeB < nodeA && nodeX < nodeB {
		nodeA -= ringSize
	}
	return nodeX > nodeA && nodeX <= nodeB
}

// BetweenLeftIncl checks if identifier is in range [a, b)
func BetweenLeftIncl(nodeX, nodeA, nodeB int) bool {
	if nodeB < nodeA && nodeX > nodeB {
		nodeB += ringSize
	} else if nodeB < nodeA && nodeX < nodeB {
		nodeA -= ringSize
	}
	return nodeX >= nodeA && nodeX < nodeB
}

// Hash provides the SHA-1 hashing required to get the identifiers for nodes and keys
func Hash(key string) int {
	hash := sha1.New()
	hash.Write([]byte(key))
	result := hash.Sum(nil)
	return int(binary.BigEndian.Uint64(result) % ringSize)
}

// PrintNode prints the node info in a formatted way
func (node *Node) PrintNode() {
	print := "=========================================================\n"
	print += "Identifier: " + strconv.Itoa(node.Identifier) + "\n"
	print += "IP: " + node.IP + "\n"
	if node.predecessor == nil {
		print += "Predecessor: nil \n"
	} else {
		print += "Predecessor: " + strconv.Itoa(node.predecessor.Identifier) + "\n"
	}
	print += "Successor: " + strconv.Itoa(node.successorList[0].Identifier) + "\n"
	print += "Successor List: "
	for _, successor := range node.successorList {
		if successor != nil {
			print += strconv.Itoa(successor.Identifier) + ", "
		}
	}
	print += "\nFinger Table: "
	for _, finger := range node.fingerTable {
		if finger != nil {
			print += strconv.Itoa(finger.Identifier) + ", "
		}
	}
	print += "\nHash Table:\n"
	for key, value := range node.hashTable {
		keyIdentifier := Hash(key)
		print += "\t" + key + " (" + strconv.Itoa(keyIdentifier) + "): " + value + "\n"
	}
	print += "\n========================================================="

	fmt.Println(print)
}

/*
	Functions to scan ip in network
*/
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	//fmt.Println(localAddr.IP)

	return localAddr.IP.String()
}

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

//  http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

type Pong struct {
	Ip    string
	Alive bool
}

func ping(pingChan <-chan string, pongChan chan<- Pong) {
	for ip := range pingChan {
		_, err := exec.Command("ping", "-c1", "-t1", ip).Output()
		var alive bool
		if err != nil {
			alive = false
		} else {
			alive = true
		}
		pongChan <- Pong{Ip: ip, Alive: alive}
	}
}

func receivePong(pongNum int, pongChan <-chan Pong, doneChan chan<- []Pong) {
	var alives []Pong
	for i := 0; i < pongNum; i++ {
		pong := <-pongChan
		//fmt.Println("received:", pong)

		if pong.Alive {
			alives = append(alives, pong)
		}
	}
	doneChan <- alives
}

func NetworkIP() (string, []string) {
	fmt.Println("Searching for IP of nodes in network ... ...")

	basicIP := GetOutboundIP()
	myIP := basicIP + "/24"
	fmt.Println(myIP)
	hosts, _ := Hosts(myIP)
	concurrentMax := 100
	pingChan := make(chan string, concurrentMax)
	pongChan := make(chan Pong, len(hosts))
	doneChan := make(chan []Pong)

	for i := 0; i < concurrentMax; i++ {
		go ping(pingChan, pongChan)
	}

	go receivePong(len(hosts), pongChan, doneChan)

	for _, ip := range hosts {
		pingChan <- ip
		//fmt.Println("sent: " + ip)
	}

	alives := <-doneChan

	var ipSlice []string

	for _, addr := range alives {
		if addr.Ip != basicIP {
			ipSlice = append(ipSlice, addr.Ip)
		}
	}

	fmt.Println("Search completed!")

	return basicIP, ipSlice
}
