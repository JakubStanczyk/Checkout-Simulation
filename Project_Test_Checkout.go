package main

import (
	"fmt"
	"sync"
	"strconv"
	"time"
)

//////////////////////////////////////////////////////////////

type subject interface {
    register(Observer observer)bool
    deregister(Observer observer)
    // notify()
}

type checkout struct {
  observerList []observer
  name string
  queueMaxLength int
	// checkoutType int
	mux sync.Mutex
}

func newCheckout(name string, queueMaxLength int) *checkout {
    return &checkout{
        name: name,
        queueMaxLength: queueMaxLength,
    }
}

func (i *checkout) register(o observer) bool {
  i.mux.Lock()
  if len(i.observerList) == i.queueMaxLength {
    i.mux.Unlock()
		// checkout queue is full
    return false
  } else {
		fmt.Println(o.getID() + " Added")
    i.observerList = append(i.observerList, o)
  }
  i.mux.Unlock()
  return true
}

// rename to leaveCheckout? or customerLeaves? or customerRemove?
func (i *checkout) deregister() {
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
	}else if observerListLength == 1 {
		return observerList[:observerListLength-1]
	}
	return observerList
}

// This needs to become a timed removal of the first item in checkout
// rename to processCustomers?
func (i *checkout) openCheckout() {
	// need a wait function that waits based on itemcount times checkout speed
		for{
			if len(i.observerList) > 0 {
				i.observerList[0].update()
				i.deregister()
			}else {
				time.Sleep(1 * time.Second)
				fmt.Println("Waiting for Customer")
			}
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

func (c *customerAgent) activateCustomer(man *manager) {
	for {
			if(man.lookingForQueue(c)) {
				fmt.Println("True " + c.getID())
				break
			}else {


				//fmt.Println("False " + c.getID())

				// will need timer here
				//break
			}
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////<--------------------------------------------

type customerFeedbackChannels struct {

		lostCustomers chan int
		waitTime chan int
		productsProcessed chan int
}

// storeReportedData? || customerFeedback
type customerFeedback struct {
	lostCustomers int
	waitTime int
	numWaited int
	productsProcessedTotal int
	numCustomersWProduct int
}



type manager struct {
	checkouts [] *checkout
	checkoutTypes []int
	customerFeedbackChannels customerFeedbackChannels

}


// what does manager need from user? == number of chekouts (1 to 8) ; types of checkouts?

// total wait time for each customer   &&   average customer wait time
// total products processed  &&  average products per trolley
//
// total utilization for each checkout  ///// time spent with no customers?
// average checkout utilisation

// The number of lost customers (Customers will leave the store if they need to join a queue more than six deep)

func newManager(checkouts [] *checkout) *manager {
    return &manager{
        checkouts: checkouts,
				 // checkoutTypes []int
        // checkoutTypes: checkoutTypes,
    }
}


// TODO: Use all checkouts
func (m *manager) lookingForQueue(customer *customerAgent) bool {

	/// try to assign customer to a queue, if unable to, return false

		// for i range := m.checkouts {
			// Need if statement..... so that we do not queue in more than one
		//fmt.Println(m.checkouts[0].register(customer))
		// }
		return m.checkouts[0].register(customer)
}


//////////////////////////////////////////////////////////////////////////

func main() {

	//TODO: Get input from user, and with that input create weather agent and managerAgent


	// hardcode new manager for testing purposes
	checkouts := []*checkout {newCheckout("Checkout one", 5), newCheckout("Checkout two", 5)}
	manager := newManager(checkouts)

  // fmt.Println(manager.checkouts[0].queueMaxLength)
	// fmt.Println(manager.checkouts[1].queueMaxLength)
	// fmt.Println(len(manager.checkouts))

 	go manager.checkouts[0].openCheckout()

	// Temporary weather agent construct //////////////////////////////
	for i:=0 ;i < 20; i++ {

		customer := &customerAgent{id: "a"+ strconv.Itoa(i)}
		go customer.activateCustomer(manager)
	}


	// Output to show that program is not frozen.
	epochs := 10

	for i:=0; i<epochs; i++ {
		time.Sleep(1 * time.Second)
		fmt.Println(".")
	}
}

