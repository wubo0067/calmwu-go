/*
 * @Author: calmwu
 * @Date: 2020-06-07 21:28:08
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-08-15 19:43:50
 */

package utils

import "math"

// Define struct
type numberResult struct {
	number  int64
	isPrime bool
}

// Define functions

// IsPrime validate N number is prime
func IsPrime(n int64) bool {
	var i, limit int64
	if n <= 1 {
		return false
	}

	if n == 2 {
		return true
	}

	if math.Mod(float64(n), 2) == 0 {
		return false
	}

	limit = int64(math.Ceil(math.Sqrt(float64(n))))
	for i = 3; i <= limit; i += 2 {
		if math.Mod(float64(n), float64(i)) == 0 {
			return false
		}
	}
	return true
}

// createNrAndValidate: Receive number and validate if is prime, send channel this same
func createNrAndValidate(n int64, c chan numberResult) {
	result := new(numberResult)
	result.number = n
	result.isPrime = IsPrime(n)
	c <- *result
}

func initGoCalculations(min int64, max int64, c chan numberResult) {
	var i int64
	for i = min; i <= max; i++ {
		go createNrAndValidate(i, c)
	}
}

func PrimesInRange(min int64, max int64) (primeArr []int64) {

	// Create channels and defer close
	c := make(chan numberResult)
	defer close(c)

	// Execute N goroutines in range number
	go initGoCalculations(min, max, c)

	for i := min; i <= max; i++ {
		// Receive numberResult
		r := <-c
		if r.isPrime {
			primeArr = append(primeArr, r.number)
		}
	}
	return
}
