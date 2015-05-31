package pike

import (
	"fmt"
	"testing"
)

func TestPike(t *testing.T) {
	fmt.Println("wordSize:", wordSize)
	ctx := NewCtx(1234)
	data := make([]byte, 1024*1024)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}
	ctx.Codec(data)
	//fmt.Println("ciphertext:", string(data), len(data))
	ctx1 := NewCtx(1234)
	ctx1.Codec(data)
	for i := 0; i < 1024*1024; i++ {
		if data[i] != byte(i%256) {
			t.Error("解码错误")
		}
	}
}

func BenchmarkPike(b *testing.B) {
	ctx := NewCtx(1234)
	data := make([]byte, 200)
	for i := 0; i < b.N; i++ {
		ctx.Codec(data)
	}
}

func TestCrack(t *testing.T) {
	plaintext2 := []byte("HISLINEISSECURE")
	ctx2 := NewCtx(1234)
	ctx2.Codec(plaintext2)
	fmt.Println(plaintext2)

	plaintext1 := []byte("THISLINEISSECURE")
	ctx1 := NewCtx(1234)
	ctx1.Codec(plaintext1)
	fmt.Println(plaintext1)
}
