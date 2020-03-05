package cbind
//#include <stdio.h>
//#include <math.h> 
//#include <fenv.h>
import "C"
import "unsafe"

func Fegetround() int {
	n := int(C.fegetround())
	return n
}