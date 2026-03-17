package ledger

import (
	"os"
	"strings"

	"github.com/spf13/pflag"
)

// AccountRole classifies an account's role in the chart of accounts.
type AccountRole string

const (
	RoleAsset     AccountRole = "asset"
	RoleLiability AccountRole = "liability"
	RoleExpense   AccountRole = "expense"
	RoleIncome    AccountRole = "income"
	RoleEquity    AccountRole = "equity"
	RoleUnknown   AccountRole = ""
)

// AccountRoleConfig maps root account name prefixes to roles.
type AccountRoleConfig struct {
	roots map[string]AccountRole
}

var (
	flagAssetRoots     string
	flagLiabilityRoots string
	flagExpenseRoots   string
	flagIncomeRoots    string
	flagEquityRoots    string
)

func init() {
	pflag.StringVar(&flagAssetRoots, "asset-roots", "", "Comma-separated asset root account names")
	pflag.StringVar(&flagLiabilityRoots, "liability-roots", "", "Comma-separated liability root account names")
	pflag.StringVar(&flagExpenseRoots, "expense-roots", "", "Comma-separated expense root account names")
	pflag.StringVar(&flagIncomeRoots, "income-roots", "", "Comma-separated income root account names")
	pflag.StringVar(&flagEquityRoots, "equity-roots", "", "Comma-separated equity root account names")
}

func newAccountRoleConfig() *AccountRoleConfig {
	cfg := &AccountRoleConfig{roots: make(map[string]AccountRole)}

	addRoots := func(flag, envVar, defaultVal string, role AccountRole) {
		val := flag
		if val == "" {
			val = os.Getenv(envVar)
		}
		if val == "" {
			val = defaultVal
		}
		for _, root := range strings.Split(val, ",") {
			root = strings.TrimSpace(root)
			if root != "" {
				cfg.roots[strings.ToLower(root)] = role
			}
		}
	}

	addRoots(flagAssetRoots, "CASHD_ASSET_ROOTS", "assets", RoleAsset)
	addRoots(flagLiabilityRoots, "CASHD_LIABILITY_ROOTS", "liability", RoleLiability)
	addRoots(flagExpenseRoots, "CASHD_EXPENSE_ROOTS", "expenses", RoleExpense)
	addRoots(flagIncomeRoots, "CASHD_INCOME_ROOTS", "income", RoleIncome)
	addRoots(flagEquityRoots, "CASHD_EQUITY_ROOTS", "", RoleEquity)

	return cfg
}

// Classify returns the role for a given root account name.
func (c *AccountRoleConfig) Classify(rootAccount string) AccountRole {
	if role, ok := c.roots[strings.ToLower(rootAccount)]; ok {
		return role
	}
	return RoleUnknown
}
