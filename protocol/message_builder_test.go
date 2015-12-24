package protocol

import (
	pb "github.com/golang/protobuf/proto"
	"testing"
)

func TestOkMessage1(t *testing.T) {
	// Check header
	msg := &Ok{Type: 234}
	if msg.Type != 234 {
		t.Fail()
	}
}

func TestOkMessage2(t *testing.T) {
	// Check header
	msg := &Ok{Type: 0}
	if msg.Type != 0 {
		t.Fail()
	}
}

func TestOkMessage3(t *testing.T) {
	// Check header
	msg := &Ok{Type: int32(M_USER_AUTH)}

	if msg.Type != int32(M_USER_AUTH) {
		t.Fail()
	}
}

func TestOkMessage4(t *testing.T) {

	msg := &Ok{Type: 234}
	data, err := pb.Marshal(msg)

	if err != nil {
		t.Error("Marshaling error:", err)
	}

	if len(data) == 0 {
		t.Error("Data Size is", len(data))
	}
}

func TestOkMessage5(t *testing.T) {

	msg := &Ok{Type: 0}
	data, err := pb.Marshal(msg)

	if err != nil {
		t.Error("Marshaling error:", err)
	}

	if len(data) != 0 {
		t.Error("Data Size is", len(data), "but should be 0")
	}
}

func TestOkMessage6(t *testing.T) {
	// Check header
	msg := NewMessage().Ok(234)
	if msg.Header.Size <= 6 {
		t.Fail()
	}
}

func TestOkMessage7(t *testing.T) {
	// Check header
	msg := NewMessage().Ok(0)
	if msg.Header.Size != 6 {
		t.Fail()
	}
}

func TestOkMessage8(t *testing.T) {
	// Check header
	msg := NewMessage().Ok(M_USER_AUTH)
	if msg.Header.Size != 6 {
		t.Fail()
	}
}

func TestErrorMessage1(t *testing.T) {
	msg := &Error{Type: int32(M_USER_CREATE_ACCOUNT), Error: E_EMAIL_EXISTS}
	if msg.Type != int32(M_USER_CREATE_ACCOUNT) || msg.Error != E_EMAIL_EXISTS {
		t.Fail()
	}
}

func TestErrorMessage2(t *testing.T) {
	msg := &Error{Type: 0, Error: 0}
	if msg.Type != 0 || msg.Error != 0 {
		t.Fail()
	}
}

func TestErrorMessage3(t *testing.T) {
	msg := &Error{Type: 0, Error: 0}
	data, err := pb.Marshal(msg)

	if err != nil {
		t.Error("Marshaling error:", err)
	}

	if len(data) != 0 {
		t.Error("Data Size is", len(data), "but should be 0")
	}
}
