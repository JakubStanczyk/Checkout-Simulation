package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

//////////////////////////////////////////////////////////////

type subject interface {
	register(Observer observer) bool
	deregister(Observer observer)
}

type checkout struct {
	observerList   []observer
	name           string
	queueMaxLength int
	mux            sync.Mutex
}

func newCheckout(name string, queueMaxLength int) *checkout {
	return &checkout{
		name:           name,
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

		i.observerList = append(i.observerList, o)
	}
	i.mux.Unlock()
	return true
}

func (i *checkout) deregister() {
	i.observerList = removeFirstElementFromslice(i.observerList)
}

func removeFirstElementFromslice(observerList []observer) []observer {
	observerListLength := len(observerList)
	// list has at least two elements
	if observerListLength > 1 {
		for i := 1; i < observerListLength; i++ {
			observerList[i-1] = observerList[i]
		}
		return observerList[:observerListLength-1]
	} else if observerListLength == 1 {
		return observerList[:observerListLength-1]
	}
	return observerList
}

func (i *checkout) openCheckout() {
	// need a wait function that waits based on itemcount times checkout speed
	//TODO:close checkout function neeted
	for {
		time.Sleep(1 * time.Second)
		if len(i.observerList) > 0 {
			i.observerList[0].update()
			i.deregister()
		} else {
			time.Sleep(1 * time.Second)

			//send message to manager
		}
	}
}

/////////////////////////////////////////////////////////////////

type observer interface {
	update()
	getID() string
}

type customerAgent struct {
	id                      string
	waitTime                float64
	patienceTime            time.Duration
	productsAmount          int
	timeStart               time.Time
	elapsedTime             time.Duration
	customerFeedbackPointer *customerFeedbackChannels
}

func (c *customerAgent) update() {
	//customer leaves shop update manager

	c.customerFeedbackPointer.productsProcessed <- c.productsAmount
	c.customerFeedbackPointer.waitTime <- c.waitTime
	c.customerFeedbackPointer.customerProcessed <- 1
}

func (c *customerAgent) getID() string {

	return c.id
}

func (c *customerAgent) genProductsAmount(userInputMin, userInputMax int) int {

	rand.Seed(time.Now().UnixNano())
	c.productsAmount = rand.Intn((userInputMax - userInputMin + 1) + userInputMin)

	return c.productsAmount
}

func (c *customerAgent) activateCustomer(man *manager) {
	//genProductsAmount() needs to be tweaked for user input
	c.timeStart = time.Now()
	c.customerFeedbackPointer = man.customerFeedbackChannels

	for {
		if man.lookingForQueue(c) {
			//found a Q

			break
		} else {
			c.patienceTime = 5 * time.Second //TODO change patience to be random gen or userin
			timePassed := time.Now()
			c.elapsedTime = (timePassed.Sub(c.timeStart))
			if c.elapsedTime >= c.patienceTime {
				c.waitTime = float64(c.elapsedTime)
				c.customerFeedbackPointer.lostCustomers <- 1
				println("cutomer leaves" + " " + c.getID())

				break
			}

		}
		// can add sleep to decrese the instensity of calls
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////

type customerFeedbackChannels struct {
	lostCustomers     chan int
	waitTime          chan float64
	productsProcessed chan int
	customerProcessed chan int
}

type customerFeedback struct {
	lostCustomers          int
	waitTime               float64
	numWaited              int
	productsProcessedTotal int
	numCustomersWProduct   int
}

type manager struct {
	checkouts                []*checkout
	checkoutTypes            []int
	customerFeedbackChannels *customerFeedbackChannels
	checkoutLenght           int
}

func newManager(checkouts []*checkout) *manager {
	return &manager{
		checkouts: checkouts,
		//// will need to change buffer size to match number of customers spawned
		customerFeedbackChannels: newFeedbackListeners(1000),
		checkoutLenght:           len(checkouts),

		// checkoutTypes []int
		// checkoutTypes: checkoutTypes,
	}
}

func newFeedbackListeners(bufsize int) *customerFeedbackChannels {
	return &customerFeedbackChannels{
		lostCustomers:     make(chan int, bufsize),
		waitTime:          make(chan float64, bufsize),
		productsProcessed: make(chan int, bufsize),
	}
}

func newFeedbackTable() *customerFeedback {
	return &customerFeedback{
		lostCustomers:          0,
		waitTime:               0,
		numWaited:              0,
		productsProcessedTotal: 0,
		numCustomersWProduct:   0,
	}
}

func (m *manager) lookingForQueue(customer *customerAgent) bool {
	for i := 0; i < m.checkoutLenght; i++ {
		//need another if for types of checkouts
		if m.checkouts[i].register(customer) == true {
			return true
		}
	}
	return false
}

//////////////////////////////////////////////////////////////////////////

type weather struct {
	customerGeneratorRate int
	stop                  bool
}

/// int needs to be between 0 - 60
func newWeatherAgent(rate int) *weather {
	return &weather{
		customerGeneratorRate: rate,
		stop:                  false,
	}
}

func (weather *weather) generateCustomers(manager *manager) {

	/// does a 0 rate mod mean no customers?
	if weather.customerGeneratorRate == 0 {
		weather.stop = true
	}

	i := 0

	rate := 20 + weather.customerGeneratorRate

	for {
		if weather.stop {
			break
		}

		i = i + 1
		time.Sleep(time.Duration(rate) * time.Millisecond)
		customer := &customerAgent{id: "a" + strconv.Itoa(i)}
		go customer.activateCustomer(manager)
	}
}

func (weather *weather) endDay() {
	weather.stop = false
}

//////////////////////////////////////////////////////////////////////////

func main() {

	//TODO: Get input from user, and with that input create weather agent and managerAgent

	//TODO: define a value for max length of feedback channels, that corresponds to the number of Customers
	//TODO: make feedback channel for checkout waitTime

	// hardcode new manager for testing purposes
	checkouts := []*checkout{newCheckout("Checkout one", 5), newCheckout("Checkout two", 5), newCheckout("Checkout three", 5), newCheckout("Checkout four", 5), newCheckout("Checkout five", 5)}
	manager := newManager(checkouts)

	weather := newWeatherAgent(1)

	for i := 0; i < manager.checkoutLenght; i++ {
		go manager.checkouts[i].openCheckout()
	}

	go weather.generateCustomers(manager)

	epochs := 10 // duration of simulation
	for i := 0; i < epochs; i++ {
		time.Sleep(1 * time.Second)
		fmt.Println(".") // Output to show that program is not frozen.
	}

	weather.endDay()

	println("store closed")
	close(manager.customerFeedbackChannels.lostCustomers)
	lostCustomersResults := 0

	for i := range manager.customerFeedbackChannels.lostCustomers {
		lostCustomersResults = lostCustomersResults + i
	}
	println(lostCustomersResults)
}
