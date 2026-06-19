package main

import (
	"fmt"
	"math/rand"
	"sync"
)

// =====================================================================
// PART 1 — MAP FUNDAMENTALS
// make, set, read a missing key, range, delete.
// =====================================================================

type basicMapStruct struct {
	name string
	age  int
}

func basicMap() {
	basicMap := make(map[string]basicMapStruct)
	basicMap["1"] = basicMapStruct{
		name: "Karan Mali",
		age:  20,
	}
	basicMap["2"] = basicMapStruct{
		name: "Karan Mali EDITED",
		age:  22,
	}
	fmt.Println("Prinint value for key which is not exited", basicMap["3"])
	for i, v := range basicMap {
		fmt.Println(" Index", i, "value", v)
	}
	delete(basicMap, "1")
	delete(basicMap, "3")
	fmt.Println("values after deleted keys")
	for i, v := range basicMap {
		fmt.Println(" Index", i, "value", v)
	}
}

// =====================================================================
// PART 1b — ASSIGNING & UPDATING MAP VALUES
// A map stores *copies* of its values, so there are a few ways to update
// one — and one way that doesn't even compile. That compile error is the
// whole reason the pointer map exists.
// =====================================================================

func mapValueAssignment() {
	// A VALUE map: each key holds a whole basicMapStruct (a copy).
	people := map[string]basicMapStruct{
		"1": {name: "Karan", age: 20},
		"2": {name: "Mali", age: 22},
	}

	// Way 1 — replace the whole struct.
	people["3"] = basicMapStruct{name: "New", age: 30}

	// Way 2 — copy it out, modify the copy, assign it back.
	temp := people["1"]
	temp.age = 2000
	people["1"] = temp
	fmt.Println("value map after reassign:", people)

	// Way 3 — DOES NOT COMPILE:
	//     people["1"].age = 2000
	// "cannot assign to struct field people[\"1\"].age in map": a map element is
	// not addressable, so you can't mutate a single field of it in place.

	// A POINTER map: each key holds a *basicMapStruct, so you CAN mutate in place.
	ptrPeople := map[string]*basicMapStruct{
		"1": {name: "Karan", age: 20},
	}
	ptrPeople["1"].age = 2000 // works: modifying through the pointer, not the map slot
	fmt.Println("pointer map after in-place update:", *ptrPeople["1"])
}

// =====================================================================
// PART 2 — THE SHARED MODEL
// The map holds *pointers*, so every goroutine shares one BankAccount,
// not a copy.
// =====================================================================

type BankAccount struct {
	accountId int
	balance   int
	mu        sync.RWMutex
}
type bankAccountStore struct {
	mu      sync.RWMutex
	Account map[int]*BankAccount
}

func CreateBankStore(bankAccount map[int]*BankAccount) *bankAccountStore {
	return &bankAccountStore{

		Account: bankAccount,
	}
}

// =====================================================================
// PART 3 — THE CONCURRENCY CRASH
// Nothing guards the map's SHAPE. Many goroutines write new keys AND range
// the map at the same time -> "fatal error: concurrent map ...".
// Run this function and it crashes on purpose — that is the point.
// =====================================================================

func ConcurrentMapIssue(bank1 map[int]*BankAccount) {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {

		wg.Go(func() {
			// The per-account lock still guards ONE balance correctly...
			bank1[1].mu.Lock()
			bank1[1].balance = 2
			bank1[1].mu.Unlock()

			// ...but NOTHING guards the map itself. A read (range) and a write
			// (new key) on the same map from different goroutines is the bug:
			for v := range bank1 { // map READ
				fmt.Println("Value of Bank", v)
			}
			bank1[i+10] = &BankAccount{ // map WRITE
				accountId: i + 20,
				balance:   i + 30,
			}

			// Force the error fast: unsynchronized writes hammering the same keys
			// reliably trips "fatal error: concurrent map writes".
			for c := 0; c < 100_000; c++ {
				bank1[1] = &BankAccount{accountId: rand.Int(), balance: rand.Int() + 1}
				bank1[2] = &BankAccount{accountId: rand.Int(), balance: rand.Int() + 1}
			}
		})
	}
	wg.Wait()
}

// =====================================================================
// PART 4 — THE NAIVE FIX (over-locking)
// One global store lock around the balance update. Correct, but it
// serializes every goroutine, so the per-account lock buys nothing.
// =====================================================================

func ConcurrentMapIssueSolved(bankstore *bankAccountStore) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("RECOVERY", r)
		}
	}()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {

		wg.Go(func() {
			fmt.Println("taking write lock on store", i)
			bankstore.mu.Lock() //we dont need this because we are not changing the map shape and we need a lock on just one account
			fmt.Println("taking write lock on inside map")
			bankstore.Account[1].mu.Lock()
			fmt.Println("updating the balence")
			bankstore.Account[1].balance = 10
			fmt.Println("unlocking the write lock on inside map", i)
			bankstore.Account[1].mu.Unlock()
			fmt.Println("unlocking bank store map", i)
			bankstore.mu.Unlock()
			bankstore.mu.RLock()
			for v := range bankstore.Account {
				fmt.Println("taking read lock on inside map", i)
				bankstore.Account[v].mu.RLock()
				fmt.Println("READING THE VALUE FROM BANK STORE", v, " the Index of iteration is", i)
				bankstore.Account[v].mu.RUnlock()

			}
			bankstore.mu.RUnlock()
			fmt.Println("taking lock on outside map to add new value")
			bankstore.mu.Lock()
			fmt.Println("updating the value ")

			bankstore.Account[i+10] = &BankAccount{
				accountId: i + 20,
				balance:   i + 30,
			}
			fmt.Println("releasing the lock on outside map ")
			bankstore.mu.Unlock()

		})
	}
	wg.Wait()
}

// =====================================================================
// PART 5 — THE OPTIMIZED FIX
// Lock is chosen by what the operation does to the MAP:
//   - balance update : Account[k] is a LOOKUP (read) -> store RLock + per-account Lock
//   - read all       : ranging the map is a read     -> store RLock + per-account RLock
//   - add a key      : changes the SHAPE (a write)   -> store Lock (exclusive)
// =====================================================================

func ConcurrentMapIssueOptimized(bankstore *bankAccountStore) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("RECOVERY", r)
		}
	}()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Go(func() {
			// (1) Update ONE existing account's balance.
			// Account[1] is a map LOOKUP (a read) -> store RLock: shared with the other
			// balance-updaters, but mutually exclusive with the shape-change Lock in step 3.
			fmt.Println("taking read lock on store to look up account", i)
			bankstore.mu.RLock()
			acc := bankstore.Account[1] // read the pointer out of the map under RLock...
			bankstore.mu.RUnlock()      // ...then release: we hold the *BankAccount, the map is no longer needed.

			fmt.Println("taking write lock on account 1", i)
			acc.mu.Lock() // per-account write lock guards the actual mutation
			fmt.Println("updating the balence")
			acc.balance = 10
			fmt.Println("unlocking the write lock on account 1", i)
			acc.mu.Unlock()

			// (2) Read every account. Ranging the map is a read -> store RLock.
			bankstore.mu.RLock()
			for v, a := range bankstore.Account {
				fmt.Println("taking read lock on inside map", i)
				a.mu.RLock()
				fmt.Println("READING THE VALUE FROM BANK STORE", v, " the Index of iteration is", i)
				_ = a.balance
				a.mu.RUnlock()
			}
			bankstore.mu.RUnlock()

			// (3) Add a NEW key = shape change = a map WRITE -> store Lock (exclusive).
			fmt.Println("taking lock on outside map to add new value")
			bankstore.mu.Lock()
			fmt.Println("updating the value ")
			bankstore.Account[i+10] = &BankAccount{
				accountId: i + 20,
				balance:   i + 30,
			}
			fmt.Println("releasing the lock on outside map ")
			bankstore.mu.Unlock()
		})
	}
	wg.Wait()
}

// =====================================================================
// main — uncomment one demo at a time to run it.
// =====================================================================

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("RECOVERY", r)
		}
	}()

	bank1 := map[int]*BankAccount{
		1: {
			accountId: 1,
			balance:   100,
		},
		2: {
			accountId: 2,
			balance:   200,
		},
	}

	// PART 1 : basicMap()
	// PART 1b: mapValueAssignment()
	// PART 3 : ConcurrentMapIssue(bank1)        // crashes on purpose
	// PART 5 : ConcurrentMapIssueOptimized(CreateBankStore(bank1))

	newBankStore := CreateBankStore(bank1)

	ConcurrentMapIssueSolved(newBankStore) // PART 4
}
