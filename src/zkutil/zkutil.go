package main

import (
	"flag"
	. "fmt"
)

var (
	zkhosts = flag.String("h", "", "the address of zookeeper cluster, e.g 172.16.130.1:2181,172.16.130.2:2181")
	command = flag.String("c", "", "the command you can execute, e.g get | list")
	path	= flag.String("p", "", "the path of zookeeper node")
)

func main() {
	flag.Parse()

	if *zkhosts == "" || *command == "" || *path == "" {
		Println("invalid option")
		return
	}

	switch *command {
	case "get":
		getNode(*path)
	case "list":
		listNode(*path)
	default:
		Println("invalid command")
	}
}

func getNode(path string) {
	zkinstance, err := GetZKInstance(*zkhosts)
	if err != nil {
		Println("getNode failed: ", err)
		return
	}
	data, err := zkinstance.ListChildren(path)
	if err != nil {
		Println("getNode failed: ", err)
		return
	}
	Println(data)
}

func listNode(path string) {
	Println("parent: ", path)
	zkinstance, err := GetZKInstance(*zkhosts)
	if err != nil {
		Println("listNode failed:", err)
		return
	}
	data, err := zkinstance.ListChildren(path)
	if err != nil {
		Println("listNode failed: ", err)
		return
	}

	for _, val := range data {
		prefix := "|--"
		Printf("%s%s\n", prefix, val)
		child_path := Sprintf("%s/%s", path, val)
		listChild(child_path, prefix)
	}
	return
}

func listChild(path, prefix string) {
	zkinstance, err := GetZKInstance(*zkhosts)
	if err != nil {
		Println("listChild failed: ", err)
		return
	}
	data, err := zkinstance.ListChildren(path)
	if err != nil {
		Println("listChild failed: ", err)
		return
	}
	prefix = Sprintf("\t%s", prefix)
	for _, val := range data {
		Printf("%s%s\n", prefix, val)
		child_path := Sprintf("%s/%s", path, val)
		listChild(child_path, prefix)
	}
}
