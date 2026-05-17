package main

import (
	"fmt"
	"unsafe"

	pb "github.com/KaranMali2001/probufs/employee-proto"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func main() {

	emp := &pb.Employee{
		Id:    0,
		Name:  nil,
		Email: "karan@12345",
	}

	fmt.Println("emp email ", emp.GetEmail(), emp.Email)
	fmt.Printf("emp Name %s %v \n", emp.GetName(), emp.Name)
	fmt.Println("ID", emp.GetId(), emp.Id)
	emp.Id = 1
	emp.Name = func(s string) *string { return &s }("Karan Mali")
	emp.Email = "karan@edited"
	fmt.Println("emp email ", emp.GetEmail(), emp.Email)
	fmt.Printf("emp Name %s %v \n", emp.GetName(), emp.Name)
	fmt.Println("ID", emp.GetId(), emp.Id)

	empList := &pb.EmployeeList{
		Employee: []*pb.Employee{emp},
	}

	fmt.Println("empList Length ", len(empList.GetEmployee()))
	for _, v := range empList.Employee {
		fmt.Println("empList email ", v.GetEmail(), v.Email)
		fmt.Printf("empList Name %s %v \n", v.GetName(), v.Name)
		fmt.Println("ID", v.GetId(), v.Id)
	}
	data, err := proto.Marshal(empList)
	if err != nil {
		fmt.Println("ERROPR", err)
	}
	fmt.Println("BINARY DATA", data)
	var diffData []byte = make([]byte, 100)
	var unmarshalData pb.EmployeeList
	fmt.Printf("before object addr: %p\n", &unmarshalData)
	fmt.Printf("before state ptr: 0x%x\n", *(*uintptr)(unsafe.Pointer(&unmarshalData)))
	fmt.Printf("before raw bytes: % x\n", unsafe.Slice((*byte)(unsafe.Pointer(&unmarshalData)), int(unsafe.Sizeof(unmarshalData))))
	err = proto.Unmarshal(diffData, &unmarshalData)
	if err != nil {
		fmt.Println("Error while decoding ", err)
	}

	fmt.Printf("after object addr: %p\n", &unmarshalData)
	fmt.Printf("after state ptr: 0x%x\n", *(*uintptr)(unsafe.Pointer(&unmarshalData)))
	fmt.Printf("after raw bytes: % x\n", unsafe.Slice((*byte)(unsafe.Pointer(&unmarshalData)), int(unsafe.Sizeof(unmarshalData))))
	statePtr := *(*uintptr)(unsafe.Pointer(&unmarshalData))
	if statePtr != 0 {
		fmt.Printf("bytes at state ptr: % x\n", unsafe.Slice((*byte)(unsafe.Pointer(statePtr)), 32))
	}
	fmt.Println("unmarshal data", prototext.Format(&unmarshalData))
}
