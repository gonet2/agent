/***********************************************************
 *
 * Diffie–Hellman key exchange
 *
 * 1. Alice and Bob agree to use a prime number p = 23 and base g = 5.
 *
 * 2. Alice chooses a secret integer a = 6, then sends Bob A = g^a mod p
 * 		A = 5^6 mod 23
 * 		A = 15,625 mod 23
 * 		A = 8
 *
 * 3. Bob chooses a secret integer b = 15, then sends Alice B = g^b mod p
 * 		B = 5^15 mod 23
 * 		B = 30,517,578,125 mod 23
 * 		B = 19
 *
 * 4. Alice computes s = B^a mod p
 * 		s = 19^6 mod 23
 * 		s = 47,045,881 mod 23
 * 		s = 2
 *
 * 5. Bob computes s = A^b mod p
 *	 	s = 8^15 mod 23
 * 		s = 35,184,372,088,832 mod 23
 * 		s = 2
 *
 * 6. Alice and Bob now share a secret (the number 2) because 6 × 15 is the same as 15 × 6
 */
package dh

import (
	"math"
	"math/big"
	"math/rand"
	"time"
)

var (
	rng         = rand.New(rand.NewSource(time.Now().UnixNano()))
	DH1BASE     = big.NewInt(3)
	DH1PRIME, _ = big.NewInt(0).SetString("0x7FFFFFC3", 0)
	MAXINT64    = big.NewInt(math.MaxInt64)
)

// Diffie-Hellman Key-exchange
func DHExchange() (*big.Int, *big.Int) {
	SECRET := big.NewInt(0).Rand(rng, MAXINT64)
	MODPOWER := big.NewInt(0).Exp(DH1BASE, SECRET, DH1PRIME)
	return SECRET, MODPOWER
}

func DHKey(SECRET, MODPOWER *big.Int) *big.Int {
	KEY := big.NewInt(0).Exp(MODPOWER, SECRET, DH1PRIME)
	return KEY
}
