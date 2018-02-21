package gostruct

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestPutBytes(t *testing.T) {
	buf := []byte{0, 0, 0, 0}
	expect := []byte{0x1, 0x2, 0x3, 0x4}
	putBytes(buf, uint64(0x04030201), true, 4)
	if bytes.Compare(expect, buf) != 0 {
		t.Errorf("LE %v != %v", expect, buf)
	}
	putBytes(buf, uint64(0x01020304), false, 4)
	if bytes.Compare(expect, buf) != 0 {
		t.Errorf("BE %v != %v", expect, buf)
	}
}

func TestGetBytes(t *testing.T) {
	buf := []byte{1, 2, 3, 4}
	expect := uint64(0x04030201)
	r := getBytes(buf, true, 4)
	if r != expect {
		t.Errorf("LE %x != %x", expect, r)
	}
	expect = uint64(0x01020304)
	r = getBytes(buf, false, 4)
	if r != expect {
		t.Errorf("BE %v != %v", expect, r)
	}

}

func packAndCheck(t *testing.T, fmt string, buflen int, iargs []interface{}, ochk []interface{}) error {
	buf := make([]byte, buflen)
	err := Pack(fmt, buf, iargs...)
	if err != nil {
		return err
	}
	oargs, err := Unpack(fmt, buf)
	if len(iargs) != len(oargs) {
		t.Fatalf("Sent %d args, got back %d: (%+v) (%+v)",
			len(iargs), len(oargs), iargs, oargs)
	}
	if ochk == nil {
		ochk = iargs
	}
	for i := 0; i < len(ochk); i++ {
		it := reflect.TypeOf(ochk[i])
		ot := reflect.TypeOf(oargs[i])
		if it != ot {
			t.Fatalf("output type mismatch at elem %d: %+v != %+v", i, it, ot)
		}
		if ochk[i] != oargs[i] {
			t.Errorf("at %d, %+#v != %+#v", i, ochk[i], oargs[i])
			t.Errorf("buf: %+v", buf)
			t.Fatalf("%+#v != %+#v", ochk, oargs)
		}
	}
	return nil
}

func TestInts(t *testing.T) {
	err := packAndCheck(t, "<h>Hi<Il>Lq<Q", 256, []interface{}{
		int16(0x0100),
		uint16(0x0203),
		int32(0x04050607),
		uint32(0x0b0a0908),
		int32(0x0f0e0d0c),
		uint32(0x00010203),
		int64(0x0405060708090a0b),
		uint64(0x0c0d0e0f00010203),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFloats(t *testing.T) {
	err := packAndCheck(t, "<fd>df", 256, []interface{}{
		float32(1.82),
		float64(8675.309),
		float64(1.134),
		float32(8.0085),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStrings(t *testing.T) {
	err := packAndCheck(t, "8s32p32s24p", 256, []interface{}{
		"hello",
		"is there anybody in there",
		"just nod if you can hear me",
		"is there anyone home?",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// truncation
	err = packAndCheck(t, "8s8p8p8s", 256, []interface{}{
		"her majesty's",
		"a pretty nice girl",
		"but she doesn't have",
		"a lot to say",
	}, []interface{}{
		"her maje",
		"a prett",
		"but she",
		"a lot to",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMisc(t *testing.T) {
	err := packAndCheck(t, "<cbB?4x>?Bbc", 256, []interface{}{
		int8(1),
		int8(2),
		byte('?'),
		bool(true),
		bool(false),
		byte('%'),
		int8(3),
		int8(4),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOverUnderflow(t *testing.T) {
	emsg := "insufficient buf space for fmt"
	epack := "Pack/Unpack succeeded, but shouldn't have. err="
	buf := make([]byte, 4)
	err := Pack("Q", buf, 0)
	if err == nil || err.Error() != emsg {
		t.Fatal(epack, err)
	}
	_, err = Unpack("Q", buf)
	if err == nil || err.Error() != emsg {
		t.Fatal(epack, err)
	}
	err = Pack("8s", buf, "Kind of Blue")
	if err == nil || err.Error() != emsg {
		t.Fatal(epack, err)
	}
	_, err = Unpack("8s", buf)
	if err == nil || err.Error() != emsg {
		t.Fatal(epack, err)
	}
	err = Pack("p", buf, "Glass Houses")
	if err == nil || err.Error() != emsg {
		t.Fatal(epack, err)
	}
	err = Pack("4p", buf, "Rio")
	if err != nil {
		t.Fatal("Pascal pack failed")
	}
	_, err = Unpack("p", buf[0:1])
	if err == nil || err.Error() != emsg {
		t.Fatal(epack, err)
	}
}

func TestBadFmt(t *testing.T) {
	err := packAndCheck(t, "3}", 8, []interface{}{
		"gabba", "gabba", "gabba",
	}, nil)
	if err == nil || strings.HasPrefix(err.Error(), "unknown format type") == false {
		t.Fatal("Pack succeeded, but shouldn't have.  err=", err)
	}
}
