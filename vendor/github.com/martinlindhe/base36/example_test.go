package base36_test

import (
	"fmt"

	"github.com/martinlindhe/base36"
)

func ExampleEncode() {

	fmt.Println(base36.Encode(5481594952936519619))
	// Output: 15N9Z8L3AU4EB
}

func ExampleDecode() {

	fmt.Println(base36.Decode("15N9Z8L3AU4EB"))
	// Output: 5481594952936519619
}

func ExampleEncodeBytes() {

	fmt.Println(base36.EncodeBytes([]byte{1, 2, 3, 4}))
	// Output: A2F44
}

func ExampleDecodeBytes() {

	fmt.Println(base36.DecodeToBytes("A2F44"))
	// Output: [1 2 3 4]
}
