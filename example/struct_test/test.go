package main

import (
	"fmt"
	"unsafe"
)

type INotify interface {
	notify(event_id int) error
}

func (p Person) notify(event_id int) error {
	fmt.Printf("Person[%s] age[%d] notify event[%d]\n", p.name, p.age, event_id)
	return nil
}

type SayHelloI interface {
	sayHello()
}

type SayHelloToEarth struct {
}

func (sh2e *SayHelloToEarth) sayHello() {
	fmt.Printf("Say Hello to Earth\n")
}

type SayHelloToMoon struct {
}

func (sh2m *SayHelloToMoon) sayHello() {
	fmt.Printf("Say Hello to Moon\n")
}

type Address struct {
	Zip    string
	Street string
}

type Person struct {
	name string
	age  int
	SayHelloI
	Address
}

func testMap() {
	personMap := make(map[string]Person)
	personMap["vivi"] = Person{
		name: "vivi",
		age:  6,
		Address: Address{
			Zip:    "2323232",
			Street: "汉正街",
		},
	}
	personMap["calmwu"] = Person{
		name: "calmwu",
		age:  41,
		Address: Address{
			Zip:    "45565656",
			Street: "三阳路",
		},
	}
}

func main() {
	var notify_obj INotify = Person{
		name:      "calmwu",
		age:       38,
		SayHelloI: new(SayHelloToEarth),
	}

	real_obj, ret := notify_obj.(Person)
	if ret {
		real_obj.notify(1)
		real_obj.SayHelloI.sayHello()
	} else {
		fmt.Println("failed")
	}

	p_person := new(Person)
	p_person.age = 4
	p_person.name = "vivi"
	p_person.Zip = "sdsdsdsd"
	p_person.Street = "五福路"

	notify_obj = INotify(p_person)
	real_obj1, ret := notify_obj.(*Person)
	if ret {
		real_obj1.notify(9)
	} else {
		fmt.Println("failed")
	}

	person1 := new(Person)
	person1_name := (*string)(unsafe.Pointer(person1))
	*person1_name = "Rancher"
	person1_age := (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(person1)) + unsafe.Offsetof(person1.age)))
	*person1_age = 99
	fmt.Printf("%#v\n", person1)

	a := ^uint32(0)
	fmt.Printf("a+1:%d\n", a+1)

	var addr *Address = nil
	addr = &person1.Address
	fmt.Printf("add:%v\n", addr)

	var p1 Person
	fmt.Printf("%#v\n", p1)
	if p1.name == "" {
		fmt.Println("p1.name is empty string")
	}
}
