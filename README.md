# 💰 cashd

`cashd` is a fast and cozy interactive TUI for personal finance management.
It allows you to effortlessly track, analyze, and gain insights into your financial transactions directly from your terminal.
`cashd` currently supports ledger/hledger and CSV as data sources

<p float="left">
  <img src="https://raw.github.com/hzqtc/cashd/master/screenshots/transaction_view.png" width="400" />
  <img src="https://raw.github.com/hzqtc/cashd/master/screenshots/account_view.png" width="400" />
</p>

## 🍴 Fork Changes

This is a fork of [hzqtc/cashd](https://github.com/hzqtc/cashd) with the following additions:

- **Ledger CSV parser** — replaces the custom text parser with `ledger csv --generated`, providing more reliable output parsing
- **Configurable account roles** — map root account names to roles (asset, liability, expense, income, equity) via CLI flags or env vars
- **Equity transaction type** — supports equity accounts alongside income and expense
- **Balance-sheet transfers** — `TransferIn` and `TransferOut` types for transfers between asset/liability accounts
- **Configurable currency symbol** — `--currency-symbol` / `CASHD_CURRENCY_SYMBOL` (default: `$`)
- **Account filtering** — `--ignore-roots` / `CASHD_IGNORE_ROOTS` to hide entire account trees
- **Fixed `--ledger` flag** — lazy path resolution so the flag works reliably with env var fallbacks

## ✨ Features

- **Interactive TUI:** Navigate through your financial data with an intuitive and responsive terminal interface.
- **Multiple Views:**
  - **Transactions:** View a detailed list of all your financial transactions, with sorting and searching capabilities.
  - **Accounts:** Get an overview of your financial accounts, including balances and transaction insights.
  - **Categories:** Analyze your spending and income by category, helping you understand where your money goes.
- **Flexible Data Loading:** Supports loading financial data from various sources.
  - **Configurable CSV Parsing:** Customize how `cashd` interprets your CSV files to match your data's format.
- **Date Range Filtering:** Filter transactions by custom date ranges (weekly, monthly, quarterly, annually) to focus on specific periods.
- **Search Functionality:** Quickly find specific transactions using keywords.
- **Financial Insights:** Visualize your financial trends with time-series charts for accounts and categories.

## 🚧 Limitations

The following limitations are known:

- Only supports `Cash`, `Bank Account` and `Credit Card` as account types
- Ledger CSV output is limited to 2 postings per transaction (multi-posting transactions are expanded by `--generated`)

### 📊 Supported Data Sources

`cashd` is designed to be flexible with your financial data. Currently, it supports:

- **CSV Files:** Load transactions from a standard CSV file. `cashd` provides extensive configuration options to correctly parse your CSV data.
- **Ledger/Hledger:** Integrate seamlessly with popular plain-text accounting tools like `ledger` and `hledger` by parsing their journal files.
  - Note: `cashd` invokes `ledger csv --generated` or `hledger csv --generated` and does not read journal files directly

## ⬇️ Installation

### 🛠️ Prerequsites

- A nerd font enabled terminal
- (Optional) ledger or hledger

### 🏗️ Build from source

To build `cashd`, ensure you have Go installed (version 1.18 or higher).

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/cashd.git
    cd cashd
    ```
2.  **Build the application:**
    ```bash
    make
    ```
3.  **Installs binary to `~/.local/bin`:**
    ```bash
    make install
    ```

## 🚀 Usage

### 📖 Loading Data from Ledger/Hledger (default)

To load transactions from a Ledger or Hledger journal file, use the `--ledger` flag:

```bash
cashd --ledger path/to/your/journal.dat
```

Alternatively, you can set the `LEDGER_FILE` or `HLEDGER_FILE` environment variables:

```bash
export LEDGER_FILE=/path/to/your/journal.dat
cashd
```

### 📂 Loading Data from a CSV File

To load transactions from a CSV file, use the `--csv` flag and `--csv-config` flag:

```bash
cashd --csv path/to/your/transactions.csv --csv-config path/to/your/config.json
```

### 🧪 Generating a Sample CSV File

The `sample` directory contains `sample.csv` and `sample-csv-config.json` for testing.

```bash
cashd --csv sample/sample.csv --csv-config sample/sample-csv-config.json
```

### Search Syntax

Searching transactions is easy by pressing the `/` key from the transactions view.
By default, `cashd` matches each keyword individually in all transaction fields.
Use a keyword prefix to specify a field for matching.

- `d:` match transaction Date, also supports `>` and `<` operators
  - For example, `d:2020-04-05`, or `d:>2020 d:<2023`
- `t:` match transaction Type (`Income`, `Expense`, `Equity`, `TransferIn`, `TransferOut`)
- `a:` match transaction Account
- `c:` match transaction Category
- `m:` match transaction Amount, also supports `>` and `<` operators
  - For example, `m:600`, or `m:>2000 m:<2500`
- `p:` match transaction Description

#### OR Logic

You can use `OR` to combine multiple search queries. Each query separated by `OR` is treated as a separate group,
and a transaction will be shown if it matches any of these groups.

For example:

- `c:food OR c:groceries`: finds transactions with category "food" or "groceries"
- `c:food m:>10 OR c:groceries m:>5`: finds transactions with category "food" and amount >10 or with category "groceries" and amount >5

#### Negative keywords

A keyword can be turned into a negative keyword by adding a `-` prefix.
`-` can be combined with other keyword prefixes to perform complex search queries, for example:

- `m:>4999 t:expense -c:loan -c:tax`: find expenses that are more than $4999 and not in the "loan" or "tax" categories
- `t:income -c:salary m:>1999`: find income transactions that are more than $1999 and not from "salary"

### 💻 Command Line Flags

- `-h`, `--help`: Show help message.
- `--csv <file_path>`: Specify the path to your CSV transaction file. This flag supports:
  - multiple files separated by `,`, e.g. `--csv "sample1.csv,sample2.csv"`
  - multiple flags each specifying one file, e.g. `--csv sample1.csv --csv sample2.csv`
  - glob, e.g. `--csv "*.csv"`
- `--csv-config <file_path>`: Specify the path to your CSV configuration JSON file.
- `--ledger <file_path>`: Specify the path to your Ledger/Hledger journal file.
- `--hide-help`: Hide in-app help panel

#### Ledger Account Configuration

These flags configure how ledger accounts are classified. Each accepts a comma-separated list of root account names (case-insensitive).

| Flag | Env Var | Default |
|------|---------|---------|
| `--asset-roots` | `CASHD_ASSET_ROOTS` | `assets` |
| `--liability-roots` | `CASHD_LIABILITY_ROOTS` | `liability` |
| `--expense-roots` | `CASHD_EXPENSE_ROOTS` | `expenses` |
| `--income-roots` | `CASHD_INCOME_ROOTS` | `income` |
| `--equity-roots` | `CASHD_EQUITY_ROOTS` | *(none)* |
| `--ignore-roots` | `CASHD_IGNORE_ROOTS` | *(none)* |
| `--currency-symbol` | `CASHD_CURRENCY_SYMBOL` | `$` |

Example with German-language account names (SKR04):

```bash
cashd --ledger buchungen.dat \
  --asset-roots "Aktiva" \
  --liability-roots "Fremdkapital" \
  --expense-roots "Aufwendungen" \
  --income-roots "Erlöse" \
  --equity-roots "Eigenkapital" \
  --ignore-roots "Tools" \
  --currency-symbol "€"
```

## ⚙️ CSV Configuration File Format

The CSV configuration file is a JSON file that defines how `cashd` should parse your CSV data.
This is particularly useful if your CSV columns or data formats differ from the default expectations.

Here's an example of the structure:

```json
{
  "columns": {
    "Period": "Date",
    "Accounts": "Account",
    "Category": "Category",
    "Note": "Description",
    "USD": "Amount",
    "Income/Expense": "Type"
  },
  "date_formats": [
    "2006-01-02",
    "2006-01-02 15:04:05",
    "01/02/2006",
    "01/02/2006 15:04:05"
  ],
  "transaction_types": {
    "income": "Income",
    "inc.": "Income",
    "expense": "Expense",
    "exp.": "Expense",
    "exps.": "Expense"
  },
  "account_types": {
    "cash": "Cash",
    "bank": "Bank Account",
    "credit card": "Credit Card"
  },
  "account_type_from_name": {
    "^cash$": "Cash",
    "checking$": "Bank Account",
    "saving(s)?$": "Bank Account",
    "card$": "Credit Card"
  }
}
```

### 📝 Config Fields:

- `columns`: A map where keys are the actual column headers in your CSV file, and values are the corresponding internal `TransactionField` names (`Date`, `Type`, `AccountType`, `Account`, `Category`, `Amount`, `Description`).
- `column_indexes` (Optional): A map where keys are `TransactionField` names and values are the 0-based index of the column in your CSV. If not provided, `cashd` will attempt to infer column indexes from the `columns` mapping and the CSV header.
- `date_formats`: An array of Go time format strings that `cashd` will attempt to use when parsing the `Date` column. The first format that successfully parses the date will be used.
- `transaction_types`: A map where keys are string values found in your CSV's "Type" column, and values are the internal `TransactionType` (`Income` or `Expense`). This allows `cashd` to understand various representations of income and expense in your data.
- `account_types`: A map where keys are string values found in your CSV's "AccountType" column, and values are the internal `AccountType` (`Cash`, `Bank Account`, `Credit Card`).
- `account_type_from_name`: A map where keys are regular expressions that will be matched against the `Account` name (case-insensitive), and values are the `AccountType` to assign if a match is found. This is useful for inferring account types when they are not explicitly provided in your CSV. If no match is found, it defaults to `Credit Card`.

## 🙏 Credit

This project is built using:

- [bubbletea](https://github.com/charmbracelet/bubbletea)
- [bubbles](https://github.com/charmbracelet/bubbles)
- [lipgloss](https://github.com/charmbracelet/lipgloss)
- [ntcharts](https://github.com/NimbleMarkets/ntcharts)
- [pflag](https://github.com/spf13/pflag)
