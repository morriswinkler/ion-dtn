package main

// #cgo LDFLAGS: -lbp
// #include <bp.h>
import "C"
import "fmt"

func BpAttach() int {
	return int(C.bp_attach())
}


func main() {

	fmt.Println(BpAttach())
	

}