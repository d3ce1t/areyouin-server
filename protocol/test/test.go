package main

import (
	proto "areyouin/protocol"
	//"encoding/hex"
	"fmt"
	//pb "github.com/golang/protobuf/proto"
)

func main() {
	src_msg := proto.NewMessage().UserAuthentication1("sargepl@gmail.com", "yope", "yeah")
	dst_msg := proto.Unmarshal(src_msg.Marshal())
	ok_msg := proto.NewMessage().Ok()

	fmt.Println("Version", src_msg.Header.Version, "Command", src_msg.Header.Command, "Size", src_msg.Header.Size)
	fmt.Println("Version", dst_msg.Header.Version, "Command", dst_msg.Header.Command, "Size", dst_msg.Header.Size)
	fmt.Println("Src", src_msg.Marshal())
	fmt.Println("Dst", dst_msg.Marshal())
	fmt.Println("Ok", ok_msg.Marshal())
}
