package main

// This is a simple example package that uses a template!

import (
	"fmt"
	ints "gotgo/slice(int)"
	stringtest "./test(string)"
	inttest "./test(int)"
	list "gotgo/list(int)"
	lists "gotgo/slice(gotgo/list(int).List)"
	sort "gotgo/sort(float64)"
)

func main() {
	ss := []string{"hello","world"}
	is := []int{5,4,3,2,1}
	fmt.Println("Head(ss)", stringtest.Head(ss))
	fmt.Println("Tail(is)", inttest.Tail(is))
	// Doesn't the following give you nightmares of lisp?
	fmt.Println(list.Cdr(list.Cons(1,list.Cons(2,list.Cons(3,nil)))))

	ints.Map(func (a int) int { return a*2 }, is)
	fmt.Println("is are now doubled: ", is)
	ls := []list.List{ *list.Cons(1,nil), *list.Cons(2,nil) }
	ls = lists.Append(ls, *list.Cons(3,nil))
	fmt.Println("I like lists: ", ls)

	tosort := []float64{5,4,3,2,1}
	sort.SortArray(tosort)
	if sort.AreSorted(tosort) { fmt.Println("The array is sorted!") }
}
