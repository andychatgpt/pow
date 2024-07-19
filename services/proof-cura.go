package services

//
///*
//#cgo LDFLAGS: -L./build -lsha512 -L/usr/local/cuda/lib64 -lcudart
//#include <stdint.h>
//#include <stdlib.h>  // For free()
//
//// Declaration of the C function
//void computeHashes(char** inputStrings, int* lengths, int numInputs, uint8_t* results);
//*/
//import "C"
//import (
//	"fmt"
//	"unsafe"
//)
//
//const DIGEST_SIZE = 64
//
//func ComputeSHA512Hashes(inputStrings []string) [][]byte {
//	numInputs := len(inputStrings)
//	lengths := make([]C.int, numInputs)
//	cStrings := make([]*C.char, numInputs)
//
//	for i, s := range inputStrings {
//		cStrings[i] = C.CString(s)
//		lengths[i] = C.int(len(s))
//	}
//
//	defer func() {
//		for _, s := range cStrings {
//			C.free(unsafe.Pointer(s))
//		}
//	}()
//
//	results := make([]byte, numInputs*DIGEST_SIZE)
//	C.computeHashes((**C.char)(unsafe.Pointer(&cStrings[0])), (*C.int)(unsafe.Pointer(&lengths[0])), C.int(numInputs), (*C.uint8_t)(unsafe.Pointer(&results[0])))
//
//	output := make([][]byte, numInputs)
//	for i := range output {
//		output[i] = results[i*DIGEST_SIZE : (i+1)*DIGEST_SIZE]
//	}
//
//	return output
//}
//
//func main() {
//	inputStrings := []string{"string1", "another string", "yet another string"}
//	hashes := ComputeSHA512Hashes(inputStrings)
//
//	for i, hash := range hashes {
//		fmt.Printf("Hash of \"%s\": ", inputStrings[i])
//		for _, b := range hash {
//			fmt.Printf("%02x", b)
//		}
//		fmt.Println()
//	}
//}
