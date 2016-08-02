package dh

import (
	"fmt"
	"testing"
)

func TestDH(t *testing.T) {
	X1, E1 := DHExchange()
	X2, E2 := DHExchange()

	fmt.Println("Secret 1:", X1, E1)
	fmt.Println("Secret 2:", X2, E2)

	KEY1 := DHKey(X1, E2)
	KEY2 := DHKey(X2, E1)

	fmt.Println("KEY1:", KEY1)
	fmt.Println("KEY2:", KEY2)

	if KEY1.Cmp(KEY2) != 0 {
		t.Error("Diffie-Hellman failed")
	}
}

func BenchmarkDH(b *testing.B) {
	for i := 0; i < b.N; i++ {
		X1, E1 := DHExchange()
		X2, E2 := DHExchange()

		DHKey(X1, E2)
		DHKey(X2, E1)
	}
}
