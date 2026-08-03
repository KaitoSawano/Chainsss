package node

// Konstanta Parameter Finansial & Desimal Eterbit (8 Presisi Desimal)
const (
	CoinUnit     = 100000000 // 1 Eterbit = 100,000,000 satuan terkecil
	InitialAirdrop = 10000 * CoinUnit
)

// ToDecimal mengubah satuan integer terkecil ke format float 8 desimal
func ToDecimal(amount uint64) float64 {
	return float64(amount) / float64(CoinUnit)
}

// ToUnits mengubah nilai float/koin ke satuan integer terkecil
func ToUnits(amount float64) uint64 {
	return uint64(amount * float64(CoinUnit))
}
