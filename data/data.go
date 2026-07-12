package data

import "embed"

// FS exposes the embedded JSON data files.
//go:embed cases.json transactions.json kyc_records.json flags.json
var FS embed.FS
