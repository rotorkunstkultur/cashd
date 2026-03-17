package ledger

import (
	"cashd/internal/data"
	"errors"
	"os"
	"os/exec"

	"github.com/spf13/pflag"
)

type LedgerDataSource struct{}

func getLedgerFilePath() string {
	if ledgerFileFlag != "" {
		return ledgerFileFlag
	} else if env := os.Getenv("LEDGER_FILE"); env != "" {
		return env
	} else if env := os.Getenv("HLEDGER_FILE"); env != "" {
		return env
	}
	return ""
}

var ledgerFileFlag string
var flagCurrencySymbol string

func init() {
	pflag.StringVar(&ledgerFileFlag, "ledger", "", "Ledger file path")
	pflag.StringVar(&flagCurrencySymbol, "currency-symbol", "", "Currency symbol for display (default: $)")
}

func (l LedgerDataSource) LoadTransactions() ([]*data.Transaction, error) {
	// Configure currency symbol
	symbol := flagCurrencySymbol
	if symbol == "" {
		symbol = os.Getenv("CASHD_CURRENCY_SYMBOL")
	}
	if symbol != "" {
		data.SetCurrencySymbol(symbol)
	}

	commands := []string{"ledger", "hledger"}

	for _, cmd := range commands {
		cmd := exec.Command(cmd, "-f", getLedgerFilePath(), "csv", "--generated")

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}

		if err := cmd.Start(); err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				continue
			} else {
				return nil, err
			}
		}

		// Stream output to parser
		config := newAccountRoleConfig()
		transactions, parseErr := parseJournal(stdout, config)
		// Wait for command to complete
		if err := cmd.Wait(); err != nil {
			return nil, err
		} else if parseErr != nil {
			return nil, parseErr
		} else {
			return transactions, nil
		}
	}

	return nil, os.ErrNotExist
}

func (l LedgerDataSource) Preferred() bool {
	return ledgerFileFlag != ""
}

func (l LedgerDataSource) Enabled() bool {
	return getLedgerFilePath() != ""
}
