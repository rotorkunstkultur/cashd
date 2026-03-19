package ledger

import (
	"cashd/internal/data"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

// csvPosting represents a single posting row from `ledger csv --generated`.
type csvPosting struct {
	date        time.Time
	description string
	account     string
	amount      float64
}

// parseJournal reads CSV output from `ledger csv --generated` and produces transactions.
// CSV columns: date, code, payee, account, currency, amount, cost, note
func parseJournal(reader io.ReadCloser, config *AccountRoleConfig) ([]*data.Transaction, error) {
	r := csv.NewReader(reader)
	r.FieldsPerRecord = -1 // allow variable fields

	var postings []csvPosting

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV: %w", err)
		}

		if len(record) < 6 {
			continue
		}

		dateStr := record[0]
		description := record[2]
		account := record[3]
		amountStr := record[5]

		parsedDate, err := time.ParseInLocation("2006/01/02", dateStr, time.Local)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %q: %w", dateStr, err)
		}

		amountStr = strings.ReplaceAll(amountStr, ",", "")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount %q: %w", amountStr, err)
		}

		postings = append(postings, csvPosting{
			date:        parsedDate,
			description: description,
			account:     account,
			amount:      amount,
		})
	}

	// Group postings by (date, description) to reconstruct transactions
	type txnKey struct {
		date time.Time
		desc string
	}

	var keys []txnKey
	groups := make(map[txnKey][]csvPosting)
	for _, p := range postings {
		key := txnKey{date: p.date, desc: p.description}
		if _, exists := groups[key]; !exists {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], p)
	}

	var transactions []*data.Transaction

	for _, key := range keys {
		group := groups[key]
		txns := pairPostings(group, key.date, key.desc, config)
		transactions = append(transactions, txns...)
	}

	return transactions, nil
}

// classifiedPosting holds a posting with its resolved role and sub-account parts.
type classifiedPosting struct {
	posting csvPosting
	role    AccountRole
	rest    string // everything after the root account prefix (e.g. "Reisekosten:Fahrtkosten")
}

// pairPostings implements the multi-posting pairing algorithm:
// 1. Filter out ignored root accounts
// 2. Classify each posting by role (unknown-role → P&L)
// 3. Separate into P&L+Equity (expense/income/equity/unknown) vs balance sheet (asset/liability)
// 4. If no balance sheet postings -> skip
// 5. BS-only transfers (no P&L, 2+ BS) -> TransferIn/TransferOut based on primary direction
// 6. Each P&L/equity posting paired with primary balance sheet posting (largest abs amount) -> one Transaction
func pairPostings(group []csvPosting, date time.Time, desc string, config *AccountRoleConfig) []*data.Transaction {
	var plPostings []classifiedPosting
	var bsPostings []classifiedPosting

	for _, p := range group {
		root := strings.SplitN(p.account, ":", 2)[0]

		// Skip ignored roots
		if config.IsIgnored(root) {
			continue
		}

		rest := ""
		if idx := strings.Index(p.account, ":"); idx >= 0 {
			rest = p.account[idx+1:]
		}

		role := config.Classify(root)
		cp := classifiedPosting{posting: p, role: role, rest: rest}

		switch role {
		case RoleExpense, RoleIncome, RoleEquity:
			plPostings = append(plPostings, cp)
		case RoleAsset, RoleLiability:
			bsPostings = append(bsPostings, cp)
		default:
			// Unknown-role postings treated as P&L
			plPostings = append(plPostings, cp)
		}
	}

	if len(bsPostings) == 0 {
		return nil
	}

	// Find primary balance sheet posting (largest absolute amount)
	primary := bsPostings[0]
	for _, bp := range bsPostings[1:] {
		if math.Abs(bp.posting.amount) > math.Abs(primary.posting.amount) {
			primary = bp
		}
	}

	account := primary.rest
	if account == "" {
		account = "unknown"
	}
	var accountType data.AccountType
	switch primary.role {
	case RoleAsset:
		if strings.EqualFold(account, "cash") {
			accountType = data.AcctCash
		} else {
			accountType = data.AcctBankAccount
		}
	case RoleLiability:
		accountType = data.AcctCreditCard
	}

	// BS-only transfers: no P&L postings but 2+ BS postings
	if len(plPostings) == 0 && len(bsPostings) >= 2 {
		var transactions []*data.Transaction
		for _, bp := range bsPostings {
			if bp.posting.account == primary.posting.account {
				continue
			}
			var txnType data.TransactionType
			if primary.posting.amount > 0 {
				txnType = data.TransferIn
			} else {
				txnType = data.TransferOut
			}

			category := bp.rest
			if category == "" {
				category = "unknown"
			}

			t := data.Transaction{
				Date:        date,
				Type:        txnType,
				Account:     account,
				AccountType: accountType,
				Category:    category,
				Amount:      math.Abs(bp.posting.amount),
				Description: desc,
			}
			transactions = append(transactions, &t)
		}
		return transactions
	}

	var transactions []*data.Transaction
	for _, pl := range plPostings {
		var txnType data.TransactionType
		switch pl.role {
		case RoleExpense:
			txnType = data.Expense
		case RoleIncome:
			txnType = data.Income
		case RoleEquity:
			txnType = data.Equity
		default:
			// Unknown-role postings: infer from amount sign
			if pl.posting.amount > 0 {
				txnType = data.Income
			} else {
				txnType = data.Expense
			}
		}

		category := pl.rest
		if category == "" {
			category = "unknown"
		}

		t := data.Transaction{
			Date:        date,
			Type:        txnType,
			Account:     account,
			AccountType: accountType,
			Category:    category,
			Amount:      math.Abs(pl.posting.amount),
			Description: desc,
		}
		transactions = append(transactions, &t)
	}

	return transactions
}
