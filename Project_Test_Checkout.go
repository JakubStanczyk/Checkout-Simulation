package main

import (
	"fmt"
	"sync"
	"strconv"
)

//////////////////////////////////////////////////////////////

type subject interface {
    register(Observer observer)bool
    deregister(Observer observer)
    notify()
}

type checkout struct {
	// uniqueness of customer is not the checkouts responsibility.
  // this is why map[] is not being used here
  observerList []observer
  name string
  queueMaxLength int
	mux sync.Mutex
}

// Checkout constructor
func newCheckout(name string, queueMaxLength int) *checkout {
    return &checkout{
        name: name,
        queueMaxLength: queueMaxLength,
    }
}

// rename to joinQueue? tryJoiningQueue?
func (i *checkout) register(o observer) bool {
  // single thread access
  i.mux.Lock()
  if len(i.observerList) == i.queueMaxLength {
    // checkout is full, do not add another customer agent
    i.mux.Unlock()
    return false
  } else {
    // checkout has space, add customer agent
    i.observerList = append(i.observerList, o)
  }
  i.mux.Unlock()
  return true
}


/////// if list is not empty --->> call .update() on first element and remove

// rename to leaveCheckout? or customerLeaves? or customerRemove?
func (i *checkout) deregister() {
		// first call update?
	
		i.observerList = removeFirstElementFromslice(i.observerList)
  }


func removeFirstElementFromslice(observerList []observer) []observer {
	observerListLength := len(observerList)

	// list has at least two elements
	if observerListLength > 1 {
		for i := 1; i < observerListLength ; i++ {
			observerList[i-1] = observerList[i]
		}
		return observerList[:observerListLength-1]
	// list has one element	
	}else if observerListLength == 1 {
		return observerList[:observerListLength-1]
	}
	// list was empty
	return observerList
}

// This needs to become a timed removal of the first item in checkout
// rename to processCustomers?
func (i *checkout) notify() {

  for _, observer := range i.observerList {
    observer.update()
    }
  }


/////////////////////////////////////////////////////////////////

type observer interface {
    update()
    getID() string
}

type customerAgent struct {
    id string
}

func (c *customerAgent) update() {

  //// this should report results #time #itemnumber etc.
  //// and destroy agent #AgentWorkDone
    fmt.Println("Notify Customer")
}

func (c *customerAgent) getID() string {
    return c.id
}

//////////////////////////////////////

func main() {
  checkoutOne := newCheckout("Checkout one", 5)
  //checkoutTwo := newCheckout("Checkout one", 7)
  fmt.Println(checkoutOne.queueMaxLength)
  //fmt.Println(checkoutTwo.queueMaxLength)

  // customerFirst := &customerAgent{id: "abc@gmail.com"}
  // customerSecond := &customerAgent{id: "xyz@gmail.com"}

  for x :=0 ; x<10; x++ {
    fmt.Println(checkoutOne.register(&customerAgent{id: "a"+ strconv.Itoa(x)}))
  }
	//for x :=0 ; x<10; x++ {
	//	fmt.Println(checkoutTwo.register(&customerAgent{id: "a"+ string(x)}))
	//}


	// fmt.Println(checkoutOne.observerList)
	for _, d := range checkoutOne.observerList {
		fmt.Println(d.getID())
	}
	checkoutOne.deregister()
	// fmt.Println(checkoutOne.observerList)
	for _, d := range checkoutOne.observerList {
		fmt.Print(d.getID() + " ")
	}
	fmt.Println()
	fmt.Println(len(checkoutOne.observerList))
	checkoutOne.deregister()
	// fmt.Println(checkoutOne.observerList)
	for _, d := range checkoutOne.observerList {
		fmt.Print(d.getID() + " ")
	}
	fmt.Println()
	fmt.Println(len(checkoutOne.observerList))
	// fmt.Println(checkoutTwo.observerList)


  // fmt.Println(checkoutOne.register(customerFirst))
  // fmt.Println(checkoutOne.register(customerSecond))
  // shirtItem.updateAvailability()
}
