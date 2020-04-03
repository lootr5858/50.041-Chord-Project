package main

import (
  "testing"
  "fmt"
  "chord/src/chord"
)

func Test1(t *testing.T) {
  // init User IP info
  IP := chord.GetOutboundIP()
  IP_str := fmt.Sprint(ip2Long(IP))
  ID := chord.Hash(IP_str)

  // create a node
  chord.ChordNode = &chord.Node{
    Identifier: -1,
  }
  chord.ChordNode.IP = IP
	chord.ChordNode.Identifier = ID
	go node_listen(IP)
	chord.ChordNode.CreateNodeAndJoin(nil)
	fmt.Println("Created chord network (" + IP + ") as " + fmt.Sprint(ID) + ".")
  chord.ChordNode.PrintNode()

  // remoteNode_IP :=                             //String of IP
	// remoteNode_IP_str := fmt.Sprint(ip2Long(remoteNode_IP)) //String of decimal IP
	// remoteNode_ID := chord.Hash(remoteNode_IP_str)          //Hash of decimal IP
	// remoteNode := &chord.RemoteNode{
	// 	Identifier: remoteNode_ID,
	// 	IP:         remoteNode_IP,
	// }
	// chord.ChordNode.IP = IP
	// chord.ChordNode.Identifier = ID
  //
	// go node_listen(IP)
	// chord.ChordNode.CreateNodeAndJoin(remoteNode)
  //
	// fmt.Println("remoteNode is (" + remoteNode_IP + ") " + fmt.Sprint(remoteNode_ID) + ".")
	// fmt.Println("Joined chord network (" + IP + ") as " + fmt.Sprint(ID) + ". ")

}
