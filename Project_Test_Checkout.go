package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
	"bufio"
	"os"
)

///////////////////////////////////////////////////////////////////////

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

func (i *checkout) openCheckout(){//channel *chan float64) {
	// need a wait function that waits based on itemcount times checkout speed
	//TODO:close checkout function needed

	// call to open checkout happens before customers are created,
	// wait for a second so that some customers can spawn
	time.Sleep(1 * time.Second)
	for {
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
	c.customerFeedbackPointer.customersProcessed <- 1
}

func (c *customerAgent) getID() string {
	return c.id
}


// TODO: 	//genProductsAmount() needs to be tweaked for user input###############################################
func (c *customerAgent) genProductsAmount(userInputMin, userInputMax int) int {
	rand.Seed(time.Now().UnixNano())
	c.productsAmount = rand.Intn((userInputMax - userInputMin + 1) + userInputMin)
	return c.productsAmount
}

func (c *customerAgent) activateCustomer(man *manager) {
	c.timeStart = time.Now()
	c.customerFeedbackPointer = man.customerFeedbackChannels
	c.genProductsAmount(1, 200)
	c.patienceTime = 5 * time.Second //TODO: change patience to be random gen or userin

	for {
		if man.lookingForQueue(c) {
			//println("cutomer queues " + c.getID())
			//found a Q
			break
		} else {

			timePassed := time.Now()
			c.elapsedTime = (timePassed.Sub(c.timeStart))
			c.waitTime = float64(c.elapsedTime)
			if c.elapsedTime >= c.patienceTime {
				// c.waitTime = float64(c.elapsedTime)
				c.customerFeedbackPointer.lostCustomers <- 1
				// println("cutomer leaves" + " " + c.getID())
				break
			}
		}
		// can add sleep to decrese the instensity of calls
	}
}

////////////////////////////////////////////////////////////////////////////////

type customerFeedbackChannels struct {
	lostCustomers     chan int
	waitTime          chan float64
	productsProcessed chan int
	customersProcessed chan int
}

type manager struct {
	checkouts                []*checkout
	checkoutTypes            []int
	checkoutFeedbackChannels []*chan float64
	customerFeedbackChannels *customerFeedbackChannels
	checkoutLenght           int
}

func newManager(checkouts []*checkout) *manager {
	// hard coded 8 because we have a max 8 checkouts and Go dosent like non constant variables
	channelArray := [8]chan float64{}
	for i := 0; i < len(checkouts); i++ {
		channelArray[i] = make(chan float64)
	}

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
		customersProcessed: make(chan int, bufsize),
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

///////////////////////////////////////////////////////////////////////////////

type customerFeedback struct {
	lostCustomers          int
	waitTime               float64
	productsProcessedTotal int
	numCustomersProcessed  int
}

func newFeedbackTable() *customerFeedback {
	return &customerFeedback{
		lostCustomers:          0,
		waitTime:               0.0,
		productsProcessedTotal: 0,
		numCustomersProcessed:  0,
	}
}

func (cf *customerFeedback) collectDataFromManager(m *manager){
	/// close all channels
	close(m.customerFeedbackChannels.lostCustomers)
	close(m.customerFeedbackChannels.waitTime)
	close(m.customerFeedbackChannels.productsProcessed)
	close(m.customerFeedbackChannels.customersProcessed)

	for i := range m.customerFeedbackChannels.lostCustomers {
		cf.lostCustomers = cf.lostCustomers + i
	}
	for i := range m.customerFeedbackChannels.waitTime {
		cf.waitTime = cf.waitTime + i
	}
	for i := range m.customerFeedbackChannels.productsProcessed {
		cf.productsProcessedTotal = cf.productsProcessedTotal + i
	}
	for i := range m.customerFeedbackChannels.customersProcessed {
		cf.numCustomersProcessed = cf.numCustomersProcessed + i
	}
}

func (cf *customerFeedback) calcDataAverages() (float64, float64){
	avgCustWait := 0.0
	avgProdPerT := 0.0
	if cf.numCustomersProcessed != 0{
		avgCustWait = ((cf.waitTime/float64(cf.numCustomersProcessed)/10000000))
		avgProdPerT = (float64(cf.productsProcessedTotal)/float64(cf.numCustomersProcessed))
	}

	return avgCustWait, avgProdPerT
}

func (cf *customerFeedback) printData(){

	avgCustWait, avgProdPerT := cf.calcDataAverages()

	fmt.Printf("\nTotal wait time for each customer:   %.0f \tseconds", (cf.waitTime/10000000))
	fmt.Printf("\nTotal products processed:            %d", cf.productsProcessedTotal)
	fmt.Printf("\nAverage customer wait time:          %.0f \tseconds", avgCustWait)
	fmt.Printf("\nAverage products per trolley:        %.2f", avgProdPerT)
	fmt.Printf("\nThe number of lost customers:        %d", cf.lostCustomers)
	// println("\n(Customers will leave the store if they need to join a queue more than six deep)")
	// we could implement the six deep, but instead we used a patience value.
}

///////////////////////////////////////////////////////////////////////////////

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

func (weather *weather) generateCustomers(manager *manager, wg sync.WaitGroup) {
	/// does a 0 rate mod mean no customers? We are assuming yes here.
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
		wg.Add(1)
		customer := &customerAgent{id: "a" + strconv.Itoa(i)}
		go customer.activateCustomer(manager)
	}
}

func (weather *weather) endDay() {
	weather.stop = false
}

//////////////////////////////////////////////////////////////////////////

func main() {
/////////////////// User Input Section /////////////////////////////////////
	var checkoutInput int64 = 0
  var customerInput int64 = 0
  var productInput float64 = 0.0
  var rateInput int64 = 0

	println(productInput)
	println(checkoutInput + customerInput + rateInput)

    scanner := bufio.NewScanner(os.Stdin)
    print("Enter Number of checkouts from 1 - 8:")
    for scanner.Scan() {
        input, _ := strconv.ParseInt(scanner.Text(), 10, 64)
        if input > 8 || input == 0 {
            print("Incorect Input, can only accept between 1 - 8 checkouts: ")
        } else {
            checkoutInput = input
            break
        }
    }

    print("Enter number of cusomers between 0 to 200: ")
    for scanner.Scan() {
        input, _ := strconv.ParseInt(scanner.Text(), 10, 64)
        if input > 200 {

            print("Incorrect Input, can only accept number of costumers between 0 to 200: ")
        } else {
            break
        }
    }

    print("Enter time per product at checkout from .5 to 6: ")
    for scanner.Scan() {
        input, _ := strconv.ParseFloat(scanner.Text(), 64)
        if input < 0.5 || input > 6 {
            print("Incorrect Input, can only time between .5 to 6: ")
        } else {
            break
        }
    }

    print("Enter rate of customer to be taken from 20 to 60: ")
    for scanner.Scan() {
        input, _ := strconv.ParseInt(scanner.Text(), 10, 64)
        if input < 20 || input > 60 {
            print("Incorrect Input, can only accept the rate of customers between 20 to 60: ")
        } else {
            break
        }
    }
////////////////////////////////////////////////////////////////////////////////
//checkouts := []*checkout{newCheckout("Checkout one", 5), newCheckout("Checkout two", 5), newCheckout("Checkout three", 5), newCheckout("Checkout four", 5), newCheckout("Checkout five", 5)}
	var wg sync.WaitGroup

	checkouts := [8]*checkout {}

	var i int64 = 0
	for i=0; i < checkoutInput; i++ {
		checkouts[i] = newCheckout("Checkout", 6)
	}
	println("############### Begin ####################")

  checkoutSlice := checkouts[:checkoutInput]

	manager := newManager(checkoutSlice)
	weather := newWeatherAgent(1)
		print(int(rateInput))

	for i := 0; i < manager.checkoutLenght; i++ {
		go manager.checkouts[i].openCheckout()
	}

	go weather.generateCustomers(manager, wg)

	/// 18 hours open time 6 hours closed
	epochs := 18 // duration of simulation
	for i := 0; i < epochs; i++ {
		time.Sleep(1 * time.Second)
		fmt.Println(".") // Output to show that program is not frozen.
	}
	// need the checkout printlins
	weather.endDay()
	// println("store closed")
	println("############### Gather results ####################")

	//// sometimes misses a few generatedCustomers being added to wait WaitGroup
	//// therefore we wait a second for the endDay() to finish its call before
	//// wg.Wait()
	time.Sleep(2 * time.Second)

	/// pause for the late customers to leave.
	/// #Safety
	wg.Wait()

	// will be called before printData() is called
	// fmt.Printf("\nTotal utilization for each checkout: %d", -1)
	// fmt.Printf("\nAverage checkout utilisation:        %d", 1)

	feedbackData := newFeedbackTable()
	feedbackData.collectDataFromManager(manager)
	feedbackData.printData()

}
