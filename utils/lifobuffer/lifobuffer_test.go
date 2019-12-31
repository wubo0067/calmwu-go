/*
 * @Author: calm.wu
 * @Date: 2019-12-31 14:24:20
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-12-31 14:31:37
 */

package lifobuffer

import (
	"errors"
	"io"
	"reflect"
	"testing"
)

func testLen(pfx string, should int, t *testing.T, b *Buffer) {
	if b.Len() != should {
		t.Errorf("[%s] unexpected length, expected %d, got %d",
			pfx, should, b.Len())
	}
}

func TestNewBuffer(t *testing.T) {
	b := NewBuffer(nil)
	if k := reflect.Indirect(reflect.ValueOf(b)).Kind(); k != reflect.Slice {
		t.Errorf(
			"[new buffer] unexpected underlying type, expected slice, got %v",
			k)
	}
	testLen("new buffer", 0, t, b)
	data := "suck my dick"
	b = NewBuffer([]byte(data))
	if string(*b) != data {
		t.Errorf("[new buffer] wron value, expected %s, got %s",
			data, string(*b))
	}
	testLen("new buffer", 12, t, b)
}

func TestReadWrite(t *testing.T) {
	b := NewBuffer(nil)
	data := "suck my dick"
	b.Write([]byte(data))
	testLen("read write", 12, t, b)
	p := make([]byte, 12)
	if n, err := b.Read(p); err != nil {
		t.Errorf("[read write] unexpected error '%v'", err)
	} else {
		if n != 12 {
			t.Errorf(
				"[read write] unexpencted number of bytes, expected 12, got %d",
				n)
		}
		testLen("read write", 0, t, b)
		if string(p) != data {
			t.Errorf("[read write] data not match, expected '%s', got '%s'",
				data, string(p))
		}
	}
	if n, err := b.Read(p); err == nil {
		t.Errorf("[read write] no io.EOF error")
	} else if err != io.EOF {
		t.Errorf("[read write] unexpected error '%v'", err)
	} else if n != 0 {
		t.Errorf("[read write] unexpencted number of bytes, expected 0, got %d",
			n)
	} else {
		testLen("read write", 0, t, b)
	}
	// sequantly
	b.Write([]byte(data))
	testLen("read write", 12, t, b)
	b.Write([]byte(data))
	testLen("read write", 24, t, b)
	b.Write([]byte(data))
	testLen("read write", 36, t, b)
	b.Read(p)
	testLen("read write", 24, t, b)
	b.Read(p)
	testLen("read write", 12, t, b)
	p = make([]byte, 12)
	b.Read(p)
	testLen("read write", 0, t, b)
	if string(p) != data {
		t.Errorf("[read] data not match, expected '%s', got '%s'",
			data, string(p))
	}
}

func TeasReadToNil(t *testing.T) {
	p := NewBuffer([]byte("suck my dick"))
	var y []byte
	n, err := p.Read(y)
	if err != nil {
		t.Errorf("[read to nil] unexpected err '%v'", err)
	}
	if n != 0 {
		t.Errorf("[read to nil] unexpected count, expected 0, got %d", n)
	}
}

func TestWriteTo(t *testing.T) {
	w := NewBuffer(nil)
	b := NewBuffer(nil)
	data := "suck my dick"
	b.Write([]byte(data))
	testLen("write to", 12, t, b)
	b.Write([]byte(data))
	testLen("write to", 24, t, b)
	b.Write([]byte(data))
	testLen("write to", 36, t, b)
	if n, err := b.WriteTo(w); err != nil {
		t.Errorf("[write to] unexpected error '%v'", err)
	} else if n != 36 {
		t.Errorf("[write to] unexpected number of bytes, expected 36, got %d",
			n)
	}
	testLen("write to", 0, t, b)
	p := make([]byte, 12)
	for i := 3; i > 0; i-- {
		w.Read(p)
		testLen("write to", (i-1)*12, t, w)
		if string(p) != data {
			t.Errorf("[write to] data not match, expected '%s', got '%s'",
				data, string(p))
		}
	}
}

func TestReadWriteByte(t *testing.T) {
	b := NewBuffer(nil)
	data := []byte("suck my dick")
	for i := 0; i < 12; i++ {
		if err := b.WriteByte(data[i]); err != nil {
			t.Errorf("[read write byte] unexpected error: %v", err)
		}
		testLen("read write data", i+1, t, b)
	}
	for i := 11; i >= 0; i-- {
		bt, err := b.ReadByte()
		if err != nil {
			t.Errorf("[read write byte] unexpected error: %v", err)
		}
		if bt != data[i] {
			t.Errorf(
				"[read write byte] unexpected byte, expected '%c', got '%c'",
				data[i], bt)
		}
		testLen("read write data", i, t, b)
	}
	bt, err := b.ReadByte()
	if err == nil {
		t.Errorf("[read write byte] no io.EOF error")
	}
	if err != io.EOF {
		t.Errorf(
			"[read write byte] unexpected error, expected 'io.EOF', got '%v'",
			err)
	}
	if bt != 0 {
		t.Errorf("[read write byte] unexpected byte value, expected 0, got %d",
			bt)
	}
}

func TestNext(t *testing.T) {
	b := NewBuffer(nil)
	data := []byte("suck my dick")
	b.Write(data)
	dick := b.Next(4)
	testLen("next", 12-4, t, b)
	if string(dick) != "dick" {
		t.Errorf("[next] unexpected value, expected '%s', got '%s'",
			"dick", string(dick))
	}
	suckMy := b.Next(900)
	testLen("next", 0, t, b)
	if string(suckMy) != "suck my " {
		t.Errorf("[next] unexpected value, expected '%s', got '%s'",
			"suck my ", string(suckMy))
	}
}

type BigCountWriter struct{}

func (b *BigCountWriter) Write(p []byte) (n int, err error) {
	n = len(p) + 1
	return
}

type SmallCountWriter struct{}

func (b *SmallCountWriter) Write(p []byte) (n int, err error) {
	n = len(p) - 1
	return
}

type ErrWriter struct{}

func (b *ErrWriter) Write(p []byte) (n int, err error) {
	err = errors.New("test error")
	n = 7
	return
}

func TestWriteToBig(t *testing.T) {
	b := NewBuffer(nil)
	b.Write([]byte("suck my dick"))
	bcw := new(BigCountWriter)
	n, err := b.WriteTo(bcw)
	if err == nil {
		t.Errorf("[write to big] no errors")
	}
	if err != ErrInvalidWriteCount {
		t.Errorf("[write to big] unexpected error, expected '%v', got '%v'",
			ErrInvalidWriteCount, err)
	}
	if int(n) <= b.Len() {
		t.Errorf("[write to big] unexpected byte count, expected >=%d, got %d",
			b.Len(), n)
	}
}

func TestWriteToErr(t *testing.T) {
	b := NewBuffer(nil)
	b.Write([]byte("suck my dick"))
	ew := new(ErrWriter)
	n, err := b.WriteTo(ew)
	if err == nil {
		t.Errorf("[write to err] no errors")
	}
	if err.Error() != "test error" {
		t.Errorf("[write to err] unexpected error, expected '%s', got '%v'",
			"test error", err)
	}
	if int(n) != 7 {
		t.Errorf("[write to err] unexpected byte count, expected 7, got %d", n)
	}
}

func TestWriteToSmall(t *testing.T) {
	b := NewBuffer(nil)
	b.Write([]byte("suck my dick"))
	scw := new(SmallCountWriter)
	n, err := b.WriteTo(scw)
	if err == nil {
		t.Errorf("[write to small] no errors")
	}
	if err != io.ErrShortWrite {
		t.Errorf("[write to small] unexpected error, expected '%s', got '%v'",
			io.ErrShortWrite, err)
	}
	if int(n) >= 12 {
		t.Errorf("[write to small] unexpected byte count, expected <12, got %d",
			n)
	}
}
