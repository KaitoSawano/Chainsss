// Copyright (c) 2026 AldianOkto. All rights reserved.
// Copyright (c) 2026 Eterbit Core.
// Use of this source code is governed by the Apache License.
// that can be found in the root directory of this repository.
// Project: Eterbit / Blockchain Core
//
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at. <http://www.apache.org/licenses/LICENSE-2.0>
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
