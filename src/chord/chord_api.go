package chord

import (
	"fmt"
	"strconv"

	// "errors"
	"log"
	// "net"
	"net/rpc"
	// "time"
)

var ChordNode *Node

/*
	packet contains the format of how messages for sent to and fro nodes
	packetType:	ping, pong, query, answer
	msg:		Details of the packet type
	List: 		A list with remote node type(contains ID and IP address of node)
	senderIP:	sender IP address
*/
type Packet struct {
	PacketType string
	Msg        string
	List       []*RemoteNode
	SenderIP   string
}

/*
	This is the format for rpc calls, an int, in this case is Listener needs to be there
	in order for the other node to call the Receive function in this node
*/
type Listener int

/*
	Handles the receival of message
	Different action for different packet type
	ping:	Ping checks if the node is alive, reply if alive
	pong:	Pong is a reply from a alive node, no need to do anything
	query:	Query is used for 2 things
				1. Lookup of which node the file resides in
				2. Find succesor
	answer: Answer is the reply to query(above)
	getSuccesorList:	A query to get the succesor lists from specified node
	getPredecessesor:	A query to get predecessor info from specified node
	notify:	A notification to inform specified node to check it's predecessor

	default:	If the packet type does not fall in any of the catergories then ignore

*/
func (l *Listener) Receive(payload *Packet, reply *Packet) error {
	// Check what packet type it is
	switch packetType := payload.PacketType; packetType {
	case "ping":
		fmt.Println("Receive ping from " + payload.SenderIP)
		// reply to ping
		*reply = Packet{"pong", "alive", nil, ChordNode.IP}
		return nil
	case "pong":
		fmt.Println("Receive pong from " + payload.SenderIP)
		// no action needed
		return nil
	case "query":
		fmt.Println("Receive query from " + payload.SenderIP)
		// Change the hash from string to int
		key, _ := strconv.Atoi(reply.Msg)
		// Call node to do the search
		node := handleQuery(key)
		// Form the packet for reply
		*reply = Packet{"answer", string(node.Identifier), []*RemoteNode{node}, ChordNode.IP}
		return nil
	case "answer":
		fmt.Println("File is in node ", payload.Msg)
		// no action needed
		return nil
	case "getSuccesorList":
		// Call node to return succesor list
		nodes := handleQuerySuccesorList()
		*reply = Packet{"SuccesorList", "", nodes, ChordNode.IP}
		return nil
	case "getPredecessesor":
		// Call node to return it's predecessor
		node := handleQueryPredecessor()
		*reply = Packet{"Predecessor", "", []*RemoteNode{node}, ChordNode.IP}
		return nil
	case "notify":
		// Call node to make changes if necessory
		handleQueryNotify(payload.List[0])
		return nil
	default:
		return nil
	}

}

/*
	Sends a ping request to a node to check if it is alive
	Use by nodes
	Args:
		senderIP: 	sender IP (will it be given in node.go?)
		receiverIP:	receiver IP (must be given by node.go)
*/
func ping(receiverIP string) bool {
	// try to handshake with other node
	client, err := rpc.Dial("tcp", receiverIP+":8081")
	if err != nil {
		// if handshake failed then the node is not even alive
		log.Fatal("Dialing:", err)
		return false
	}

	// Set up arguments
	payload := &Packet{"ping", "Are you alive?", nil, ChordNode.IP}
	var reply Packet

	// and make an rpc call
	err = client.Call("Listener.Receive", payload, &reply)
	if err != nil {
		log.Fatal("Connection error:", err)
		return false
	}
	fmt.Println(reply.SenderIP + " is alive. ")
	return true
}

/*
	TCP guarentees a reply so pong has not been implemented
*/
func pong() {

}

/*
	Set up query for either
		1. Node which file resides in
		2. Successor node
	Use by node
	Args:
		id:	hash of filename (key for lookup) OR
			hash of IP
	return id of node who holds the file
*/
func query(id int, closestPredIP string) *RemoteNode {
	// query closest predecessor
	// get closestPred IP
	client, err := rpc.Dial("tcp", closestPredIP+":8081")
	if err != nil {
		log.Fatal("Dialing:", err)
	}

	// set up arguments
	payload := &Packet{"query", string(id), nil, ChordNode.IP}
	var reply Packet

	// and make an rpc call
	err = client.Call("Listener.Receive", payload, &reply)
	if err != nil {
		log.Fatal("Connection error:", err)
	}

	return reply.List[0]
}

func handleQuery(id int) *RemoteNode {
	// search closest successor
	closestPred := ChordNode.findSuccessor(id)
	return closestPred
}

/*
	Have not implemented answer function because it is a recursive query
	and tcp ensures a reply
*/
func answer() {

}

/*
	Set up query for getting successor list
	Use by node
	Args:
		receiverIP:	IP of node you want to query
	return successor list
*/
func querySuccessorList(receiverIP string) []*RemoteNode {

	client, err := rpc.Dial("tcp", receiverIP+":8081")
	if err != nil {
		log.Fatal("Dialing:", err)
	}

	// set up arguments
	payload := &Packet{"getSuccesorList", "Get successor list", nil, ChordNode.IP}
	var reply Packet

	// and make an rpc call
	err = client.Call("Listener.Receive", payload, &reply)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	return reply.List
}

func handleQuerySuccesorList() []*RemoteNode {
	// Call function in node.go
	return ChordNode.successorList
}

/*
	Set up query for getting predecessor
	Use by node
	Args:
		receiverIP:	IP of node you want to query
	return predecessor
*/
func queryPredecessor(receiverIP string) []*RemoteNode {

	client, err := rpc.Dial("tcp", receiverIP+":8081")
	if err != nil {
		log.Fatal("Dialing:", err)
	}

	// set up arguments
	payload := &Packet{"getPredecessor", "Get predecessor", nil, ChordNode.IP}
	var reply Packet

	// and make an rpc call
	err = client.Call("Listener.Receive", payload, &reply)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	return reply.List
}

func handleQueryPredecessor() *RemoteNode {
	// Call node to return predecessor
	return ChordNode.predecessor
}

/*
	Set up notification
	Use by node
	Args:
		receiverIP:	IP of node you want to notify
*/
func notify(receiverIP string) {
	client, err := rpc.Dial("tcp", receiverIP+":8081")
	if err != nil {
		log.Fatal("Dialing:", err)
	}

	// set up arguments
	remoteNode := RemoteNode{ChordNode.Identifier, ChordNode.IP}
	payload := &Packet{"notify", "I am our predecessor.", []*RemoteNode{&remoteNode}, ChordNode.IP}
	var reply Packet

	// and make an rpc call
	err = client.Call("Listener.Receive", payload, &reply)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	return
}

func handleQueryNotify(sender *RemoteNode) {
	// Call node to make changes
	ChordNode.notify(sender)
}
