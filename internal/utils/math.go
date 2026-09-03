package utils

import "math/big"

// numberA == numberB
func Equal(numberA, numberB *big.Float) bool {
	return numberA.Cmp(numberB) == 0
}

// numberA > numberB
func GreaterThan(numberA, numberB *big.Float) bool {
	return numberA.Cmp(numberB) == 1
}

// numberA < numberB
func LessThan(numberA, numberB *big.Float) bool {
	return numberA.Cmp(numberB) == -1
}
