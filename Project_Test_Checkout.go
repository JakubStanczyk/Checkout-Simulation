/// @auth
/// Kristofs Flaks 15169081
/// Jakub Stanczyk 16151011
/// Jonathan Ryley 17244501

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

///////////////////////////////////////////////////////////////////////////////

type subject interface {
	register(Observer observer) bool
	deregister(Observer observer)
}

type checkout struct {
	observerList   []observer
	name           string
	queueMaxLength int
	open           bool
	cSpeed				 float64
	mux            sync.Mutex
}

func newCheckout(name string, queueMaxLength int, cSpeed float64) *checkout {
	return &checkout{
		name:           name,
		queueMaxLength: queueMaxLength,
		open:						true,
		cSpeed: 			cSpeed,
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

func (c *checkout) openCheckout(channel *chan float64) {
	// need a wait function that waits based on itemcount times checkout speed

	// call to open checkout happens before customers are created,
	// wait for a second so that some customers can spawn
	time.Sleep(1 * time.Second)

	for c.open {
		if len(c.observerList) > 0 {
			// c.observerList[0].itemcount * cSpeed
			pSpeed := float64(c.observerList[0].getItemCount()) * c.cSpeed
			time.Sleep(time.Duration(pSpeed) * time.Millisecond)
			c.observerList[0].update()
			c.deregister()
		} else {
			//send message to manager
			*channel <- 1.0
			time.Sleep(1 * time.Second)
		}
	}

	close(*channel)
}

func (i *checkout) closeCheckout() {
	i.open = false
}

/////////////////////////////////////////////////////////////////

type observer interface {
	update()
	getID() string
	getItemCount() int
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

func (c *customerAgent) getItemCount() int {
	return c.productsAmount
}

func (c *customerAgent) genProductsAmount(userInputMin, userInputMax int) int {
	rand.Seed(time.Now().UnixNano())
	c.productsAmount = rand.Intn((userInputMax - userInputMin + 1) + userInputMin)
	return c.productsAmount
}

func (c *customerAgent) activateCustomer(man *manager) {
	c.timeStart = time.Now()
	c.customerFeedbackPointer = man.customerFeedbackChannels

	c.patienceTime = 5 * time.Second //TODO: change patience to be random gen or userin
	// rate = rand.Intn((weather.customerMaxGenRat - weather.customerMinGenRate + 1) + weather.customerMinGenRate)

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
	checkoutFeedbackChannels [8]chan float64
	customerFeedbackChannels *customerFeedbackChannels
	checkoutLenght           int
}

func newManager(checkouts []*checkout) *manager {
	// hard coded 8 because we have a max 8 checkouts and Go dosent like non constant variables
	channelArray := [8]chan float64{}
	for i := 0; i < len(checkouts); i++ {
		channelArray[i] = make(chan float64, 100000) /// buffered channels for total
	}

	return &manager{
		checkouts: checkouts,
		//// will need to change buffer size to match number of customers spawned
		customerFeedbackChannels: newFeedbackListeners(1000),
		checkoutLenght:           len(checkouts),
		checkoutFeedbackChannels: channelArray,
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
		//need to refactor how types of checkouts are implemented
		// checkout 1 (index 0) is 5 items or less
		if i == 0 && customer.getItemCount() < 6 {
			if m.checkouts[i].register(customer) == true {
				return true
			}
			/// catch if there is only one checkout
		}else if i > 0 || m.checkoutLenght == 1 {
			if m.checkouts[i].register(customer) == true {
				return true
			}
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

	fmt.Printf("\nTotal wait time for each customer:   %.2f\tseconds", (cf.waitTime/10000000))
	fmt.Printf("\nAverage customer wait time:          %.2f\tseconds", avgCustWait)
	fmt.Printf("\nTotal products processed:            %d", cf.productsProcessedTotal)
	fmt.Printf("\nAverage products per trolley:        %.2f", avgProdPerT)
	fmt.Printf("\nThe number of lost customers:        %d", cf.lostCustomers)
	// println("\n(Customers will leave the store if they need to join a queue more than six deep)")
	// we could implement the six deep, but instead we used a patience value.
}

///////////////////////////////////////////////////////////////////////////////

type weather struct {
	customerMaxGenRate int
	customerMinGenRate int
	stop                  bool
}

/// int needs to be between 0 - 60
func newWeatherAgent(rateMax int, rateMin int) *weather {
	return &weather{
		customerMaxGenRate: rateMax,
		customerMinGenRate: rateMin,
		stop:                  false,
	}
}

func (weather *weather) generateCustomers(manager *manager, wg sync.WaitGroup, prodMin int, prodMax int) {
	rate := 1
	/// does a 0 rate mod mean no customers? We are assuming yes here.
	if weather.customerMaxGenRate == 0 {
		weather.stop = true
	}else{
		// higher means faster, must reduce pause rate value by input
		rate = 70 - rand.Intn((weather.customerMaxGenRate - weather.customerMinGenRate + 1) + weather.customerMinGenRate)
	}

	i := 0

	for {
		if weather.stop {
			break
		}
		i = i + 1
		time.Sleep(time.Duration(rate) * time.Millisecond)
		wg.Add(1)
		customer := &customerAgent{id: "a" + strconv.Itoa(i)}
		customer.genProductsAmount(prodMin, prodMax)
		go customer.activateCustomer(manager)
	}
}

func (weather *weather) endDay() {
	weather.stop = false
}

////////////////////////////////////////////////////////////////////////////

func main() {
/////////////////// User Input Section /////////////////////////////////////
	var checkoutInput int64 = 0
  var customerMaxEntryRate int64 = 0
	var customerMinEntryRate int64 = 0
  var checkoutMaxSpeed float64 = 0.0
	var checkoutMinSpeed float64 = 0.0
	var minProdRange int64 = 0
	var maxProdRange int64 = 0

	// the number of products for each trolley to be generated randomly within a user specified range (1 to 200).

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

    print("Enter min range of products for each trolley, 0 to 199: ")
    for scanner.Scan() {
        input, _ := strconv.ParseInt(scanner.Text(), 10, 64)
        if input > 199 {

            print("Incorrect Input, can only accept number from 0 to 199: ")
        } else {
					minProdRange = input
					break
        }
    }

		print("Enter max range of products for each trolley, 1 to 200: ")
		for scanner.Scan() {
				input, _ := strconv.ParseInt(scanner.Text(), 10, 64)
				if input > 200 && input >= 1{

						print("Incorrect Input, can only accept number from 1 to 200: ")
				} else {
					maxProdRange = input
					break
				}
		}

    print("Enter checkout min speed, from .5 to 6: ")
    for scanner.Scan() {
        input, _ := strconv.ParseFloat(scanner.Text(), 64)
        if input < 0.5 || input > 6 {
            print("Incorrect Input, can only time between .5 to 6: ")
        } else {
					checkoutMinSpeed = input
            break
        }
    }
		print("Enter checkout max speed, from .5 to 6: ")
		for scanner.Scan() {
				input, _ := strconv.ParseFloat(scanner.Text(), 64)
				if input < 0.5 || input > 6 {
						print("Incorrect Input, can only time between .5 to 6: ")
				} else {
					checkoutMaxSpeed = input
						break
				}
		}

    print("Enter maximum rate of customer spawn, from 0 to 60: ")
    for scanner.Scan() {
        input, _ := strconv.ParseInt(scanner.Text(), 10, 64)
        if input < 0 || input > 60 {
            print("Incorrect Input, can only accept number in range of 20 to 60: ")
        } else {
					customerMaxEntryRate = input
					break
        }
    }

		print("Enter minimum rate of customer spawn, from 0 to 60: ")
		for scanner.Scan() {
				input, _ := strconv.ParseInt(scanner.Text(), 10, 64)
				if input < 0 || input > 60 {
						print("Incorrect Input, can only accept number in range of 20 to 60: ")
				} else {
					customerMinEntryRate = input
					break
				}
		}
////////////////////////////////////////////////////////////////////////////////
	var wg sync.WaitGroup
	var checkoutSpeed float64 = 0.0
	checkouts := [8]*checkout {}

	var i int64 = 0
	for i=0; i < checkoutInput; i++ {
		checkoutSpeed = (checkoutMinSpeed+(rand.Float64()*(checkoutMaxSpeed - checkoutMinSpeed)))
		checkouts[i] = newCheckout(strconv.FormatInt((i + 1), 10), 6, checkoutSpeed)
	}
	println("####################### Begin #########################")

  checkoutSlice := checkouts[:checkoutInput]

	manager := newManager(checkoutSlice)

	////////////////////////////////////////////// set up checkout to be 5 items or less

	newCustEnRaMa, _ := strconv.Atoi(strconv.FormatInt(customerMaxEntryRate, 10))
	newCustEnRaMi, _ := strconv.Atoi(strconv.FormatInt(customerMinEntryRate, 10))
	weather := newWeatherAgent(newCustEnRaMa, newCustEnRaMi)

	for i := 0; i < manager.checkoutLenght; i++ {
		go manager.checkouts[i].openCheckout(&manager.checkoutFeedbackChannels[i])
	}

	newCustItMa, _ := strconv.Atoi(strconv.FormatInt(maxProdRange, 10))
	newCustItMi, _ := strconv.Atoi(strconv.FormatInt(minProdRange, 10))
	go weather.generateCustomers(manager, wg, newCustItMi, newCustItMa)

	/// 18 hours open time 6 hours closed
	epochs := 18 // duration of simulation
	println("")
	for i := 0; i < epochs; i++ {
		time.Sleep(1 * time.Second)
		fmt.Print(".") // Output to show that program is not frozen.
	}
	// need the checkout printlins
	weather.endDay()
	// println("store closed")

	//for i := 0; i < manager.checkoutLenght; i++ {
		//manager.checkouts[i].closeCheckout()
	//}

	println("\n################## Gather results #####################")

	//// sometimes misses a few generatedCustomers being added to wait WaitGroup
	//// therefore we wait a second for the endDay() to finish its call before
	//// wg.Wait()
	time.Sleep(2 * time.Second)

	/// pause for the late customers to leave.
	/// #Safety
	wg.Wait()


	////TODO: send close to checkout channel and end loop
	for i := 0; i < manager.checkoutLenght; i++ {
		manager.checkouts[i].closeCheckout()
	}

	var localTotal float64 = 0.0
	var total float64 = 0.0
	var average float64 = 0.0
	//println("Total utilization for each checkout:")
	for i := 0; i < manager.checkoutLenght; i++ {
		localTotal = 0.0
		for j := range manager.checkoutFeedbackChannels[i] {
			localTotal = localTotal + j
		}
		localTotal = (localTotal * 10)
		total = total + localTotal
		fmt.Printf("\nCheckout %s waiting for customers:     %.0f\tseconds", manager.checkouts[i].name ,localTotal)
	}
	//for each checkout
	average = total/float64(manager.checkoutLenght)
	fmt.Printf("\nTotal checkout utilization:          %.0f\tseconds", total)
	fmt.Printf("\nAverage checkout utilisation:        %.0f\tseconds", average)

	feedbackData := newFeedbackTable()
	feedbackData.collectDataFromManager(manager)
	feedbackData.printData()

}


