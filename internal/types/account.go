package types

import "fmt"

// Account represents an Ethereum account
type Account struct {
	Address string
	Nonce   uint64
	Balance uint64
}

func NewAccount(address string, balance uint64) *Account {
	return &Account{
		Address: address,
		Nonce:   0,
		Balance: balance,
	}
}

// check if the account has enough balance
func (a *Account) CanPay(amount uint64) bool {
	return a.Balance >= amount
}

// remove the amount from balance
func (a *Account) Deduct(amount uint64) error {
	if !a.CanPay(amount) {
		return fmt.Errorf("innufficient balance: have %d, need %d", a.Balance, amount)
	}

	a.Balance -= amount
	return nil
}

// add amount to balance
func (a *Account) Add(amount uint64) {
	a.Balance += amount
}

func (a *Account) IncrementNonce() {
	a.Nonce++
}

// Satate represents teh global state (account)
type State struct {
	Accounts map[string]*Account
}

func NewState() *State {
	return &State{
		Accounts: make(map[string]*Account),
	}
}

func (s *State) GetAccount(address string) *Account {
	if acc, exists := s.Accounts[address]; exists {
		return acc
	}

	acc := NewAccount(address, 0)
	s.Accounts[address] = acc
	return acc
}

func (s *State) SetAccount(address string, account *Account) {
	s.Accounts[address] = account
}

func (s *State) GetBalance(address string) uint64 {
	return s.GetAccount(address).Balance
}

func (s *State) GetNonce(address string) uint64 {
	return s.GetAccount(address).Nonce
}
