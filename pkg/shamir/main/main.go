package main

import (
	"fmt"

	"go.dedis.ch/kyber/v3/suites"
)

func main() {
	s := suites.MustFind("Ed25519")
	x := s.Scalar().Zero()
	fmt.Println(x)

}
