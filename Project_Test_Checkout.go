package main

import (
	"fmt"
	"sync"
)

////////////////////////////////////////

type subject interface {
    register(Observer observer)
    deregister(Observer observer)
    notify()
}

type checkout struct {
	// queue map[string]bool
  observerList []observer
  name string
	mux sync.Mutex
}

func newCheckout(name string) *checkout {
    return &checkout{
        name: name,
    }
}

// func (i *item) updateAvailability() { fmt.Printf("Item %s is now in stock\n", i.name) i.inStock = true i.notifyAll() }
func (i *checkout) register(o observer) {
  i.mux.Lock()
  // add if and return bool
  i.observerList = append(i.observerList, o)
  i.mux.Unlock()
}
// func (i *item) deregister(o observer) { i.observerList = removeFromslice(i.observerList, o) }
// func (i *item) notifyAll() {    for _, observer := range i.observerList { observer.update(i.name) } }



//////////////////////////////////////

type observer interface {
    update(string)
    getID() string
}

type customerAgent struct {
    id string
}

func (c *customerAgent) update(itemName string) {
    fmt.Println("Notify Customer")
}

func (c *customerAgent) getID() string {
    return c.id
}

//////////////////////////////////////

func main() {
  checkoutOne := newCheckout("Checkout one")
  customerFirst := &customerAgent{id: "abc@gmail.com"}
  customerSecond := &customerAgent{id: "xyz@gmail.com"}
  checkoutOne.register(customerFirst)
  checkoutOne.register(customerSecond)
  // shirtItem.updateAvailability()
}
