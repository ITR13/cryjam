package main

import (
	"fmt"

	"github.com/ITR13/cryjam/simpleBruteForcer"
)

func main() {
	fmt.Println(bruteforce.Affirm([][4]string{{"-ss", "e.exe cat", "", ""},
		{"-alphabet", "abcdefg", "1", "200"}},
		[2]string{}, false, "\n", 12, 500))

	br := bruteforce.GetBruteForcer([][4]string{{"-ss", "e.exe cat", "", ""},
		{"-alphabet", "abcdefg", "1", "200"}},
		[2]string{}, false, "\n", 12, 50)
	br.BruteForce(false)
	br.Close()
}
