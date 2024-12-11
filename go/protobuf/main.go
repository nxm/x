package main

import (
	"git.jakub.app/jakub/X/protobuf/msg"
	"github.com/golang/protobuf/proto"
	"log"
)

func main() {
	// create a new "message"
	msg1 := &msg.Msg{
		Key:   "Hello Protocol Buffers",
		Value: []int64{1, 2, 3, 4},
	}

	// structure to bytes
	data, err := proto.Marshal(msg1)
	if err != nil {
		log.Fatal("marshaling error: ", err)
		return
	}

	// how much memory does it take?
	log.Printf("data length: %d", len(data))

	// bytes into the structure
	msg2 := new(msg.Msg)
	err = proto.Unmarshal(data, msg2)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}

	// now both structures must be equal
	if msg1.Key != msg2.Key {
		log.Printf("unexpected value, expected '%s', got '%s'", msg1.Key, msg2.Key)
	}

	for i := 0; i < 4; i++ {
		if msg1.Value[i] != msg2.Value[i] {
			log.Printf("unexpected value, expected %d, got %d", msg1.Value[i], msg2.Value[i])
		}
	}

	log.Println("Done")
}
