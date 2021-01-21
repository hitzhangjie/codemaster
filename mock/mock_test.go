package mock

import (
	"fmt"
)

type Vehical interface {
	Run()
	String() string
}

type Car struct {
	wheels int
}

func (c *Car) Run() {}

func (c *Car) String() string {
	return fmt.Sprintf("this car has %d wheels", c.wheels)
}


