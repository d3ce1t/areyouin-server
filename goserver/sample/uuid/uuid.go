package main

import (
	"fmt"
	"github.com/twinj/uuid"
)

func main() {
	uuid.SwitchFormat(uuid.CleanHyphen)

	uP, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	u4 := uuid.NewV4()
	fmt.Printf("version %d variant %x: %s\n", uP.Version(), uP.Variant(), uP)
	fmt.Printf("version %d variant %x: %s\n", u4.Version(), u4.Variant(), u4)

	u4_copy := uuid.New(u4.Bytes())
	fmt.Printf("version %d variant %x: %s\n", u4_copy.Version(), u4_copy.Variant(), u4_copy)

	bytes := []byte{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150, 160}

	inv_uuid := uuid.New(bytes)

	if uuid.Equal(uP, u4) {
		fmt.Printf("Will never happen")
	}

	fmt.Println(u4.String())
	fmt.Println(inv_uuid)
}
