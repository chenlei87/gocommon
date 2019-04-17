package main

import (
	"fmt"

	"git.topvdn.com/NuclearPower/DGraphForPOGA/compontent/mygoleveldb"
)

func main() {
	pmap := mypersistmap.NewPersistMap("persistmap.file0")

	//pmap.Set(1, 1)

	fmt.Println("pmap len", pmap.Len())
}
