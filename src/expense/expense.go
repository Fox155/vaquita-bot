package expense

import (
	"fmt"
	"math"
	"sort"

	"github.com/google/uuid"
)

var MapByGroup map[int64][]Expense = make(map[int64][]Expense)

type Expense struct {
	ID        uuid.UUID
	Name      string
	Amount    float64
	PayerName string
}

func InitGroup(group int64) {
	newGroup := []Expense{
		{
			PayerName: "Fox_SL",
			Name:      "Mate",
			Amount:    50,
		},
		{
			PayerName: "Fox_SL",
			Name:      "Gato",
			Amount:    50,
		},
		{
			PayerName: "Ismael",
			Name:      "Hielo",
			Amount:    50,
		},
		{
			PayerName: "mliezun",
			Name:      "tuki",
			Amount:    150,
		},
		{
			PayerName: "mliezun",
			Name:      "tuki",
			Amount:    50,
		},
	}

	MapByGroup[group] = newGroup
}

func CleanGroup(group int64) {
	MapByGroup[group] = make([]Expense, 0)
}

func NewExpense(group int64, amount float64, payerName, descrip string) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	t := Expense{
		ID:        id,
		Amount:    amount,
		PayerName: payerName,
		Name:      descrip,
	}

	MapByGroup[group] = append(MapByGroup[group], t)

	return nil
}

func (expense Expense) Balance(group int64) map[string]float64 {
	participants := []string{}
	for _, exp := range MapByGroup[group] {
		participants = append(participants, exp.PayerName)
	}
	participants = removeDuplicateStr(participants)

	balance := make(map[string]float64)
	balance[expense.PayerName] = expense.Amount
	for _, participant := range participants {
		balance[participant] += -expense.Amount / float64(len(participants))
	}
	return balance
}

func GetTukiFullBalance(group int64) (fullBalance map[string]float64, err error) {
	expenses := MapByGroup[group]
	fullBalance = make(map[string]float64)
	for _, participant := range MapByGroup[group] {
		fullBalance[participant.PayerName] = 0
	}
	for _, expense := range expenses {
		balanceExpense := expense.Balance(group)
		for k, v := range balanceExpense {
			fullBalance[k] += v
		}
	}
	if err != nil {
		return
	}
	return
}

func GetFullBalance(group int64) (fullBalance map[string]float64, err error) {
	fullBalance = make(map[string]float64)
	for _, expense := range MapByGroup[group] {
		if _, ok := fullBalance[expense.PayerName]; !ok {
			fullBalance[expense.PayerName] = expense.Amount
		} else {
			fullBalance[expense.PayerName] = fullBalance[expense.PayerName] + expense.Amount
		}

	}

	return
}

func GetTotalBalance(group int64) float64 {
	var full float64

	for _, expense := range MapByGroup[group] {
		full += expense.Amount
	}

	return full
}

type Debt struct {
	Debtor   string
	Creditor string
	Amount   float64
}

func GetDebts(group int64) (debts []Debt, err error) {
	debts = make([]Debt, 0)
	balance, err := GetTukiFullBalance(group)
	if err != nil {
		return
	}

	type kv struct {
		Key   string
		Value float64
	}
	var sortedBalance []kv
	for k, v := range balance {
		sortedBalance = append(sortedBalance, kv{k, v})
	}

	sort.Slice(sortedBalance, func(i, j int) bool {
		return sortedBalance[i].Value < sortedBalance[j].Value
	})

	i := 0
	j := len(sortedBalance) - 1
	var debt float64
	for i < j {
		debt = math.Min(-(sortedBalance[i].Value), math.Abs(sortedBalance[j].Value))
		fmt.Println(sortedBalance[i].Key, sortedBalance[i].Value)
		fmt.Println(sortedBalance[j].Key, sortedBalance[j].Value)
		fmt.Println(debt)

		sortedBalance[i].Value += debt
		sortedBalance[j].Value -= debt

		debts = append(debts, Debt{
			Debtor:   sortedBalance[i].Key,
			Creditor: sortedBalance[j].Key,
			Amount:   debt,
		})

		if sortedBalance[i].Value == 0 {
			i++
		}
		if sortedBalance[j].Value == 0 {
			j--
		}
	}
	return
}

func GetExpenses(group int64) []Expense {
	expenses := MapByGroup[group]

	sort.Slice(expenses, func(i, j int) bool {
		return expenses[i].PayerName < expenses[j].PayerName
	})

	return expenses
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
