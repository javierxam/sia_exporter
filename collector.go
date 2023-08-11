package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.sia.tech/siad/modules"
	sia "go.sia.tech/siad/node/api/client"
	"gitlab.com/NebulousLabs/errors"
	)

var (
	// ErrAPICallNotRecognized is returned by API client calls made to modules that
	// are not yet loaded.
	ErrAPICallNotRecognized = errors.New("API call not recognized")

	
	// Wallet Metrics
	walletModuleLoaded = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wallet_module_loaded", Help: "Is the wallet module loaded. 0=not loaded.  1=loaded"})
	walletLocked = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wallet_locked", Help: "Is the wallet locked. 0=not locked.  1=locked"})
	walletConfirmedSiacoinBalance = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wallet_confirmed_siacoin_balance", Help: "Wallet confirmed Siacoin balance (Siacoins)"})
	
	// Host Metrics
	hostAcceptingContracts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_accepting_contracts", Help: "Is the host accepting contracts 0=no, 1=yes"})
	hostWindowSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_window_size", Help: "Window Size in hours"})
	hostCollateral = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_collateral", Help: "Host Collateral in Siacoins"})
	hostCollateralBudget = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_collateral_budget", Help: "Host Collateral budget in Siacoins"})
	hostMaxCollateral = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_max_collateral", Help: "Max collateral per contract"})
	hostContractCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_contract_count", Help: "number of host contracts"})
	hostTotalStorage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_total_storage", Help: "total amount of storage available on the host in bytes"})
	hostRemainingStorage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_remaining_storage", Help: "amount of storage remaining on the host in bytes"})
	hostLockedCollateral = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_locked_collateral", Help: "Locked collateral"})
	hostIngressRevenue = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_ingress_potential", Help: "Ingress potential revenue"})
	hostEgressRevenue = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_egress_potential", Help: "Egress potential revenue"})
	hostStorageRevenue = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_storage_potential", Help: "Storage potential revenue"})
	)

const (
	moduleNotReadyStatus = "Module not loaded or still starting up"
)

func hostMetrics(sc *sia.Client) {
	hg  := sc.HostGet()
	sg := sc.HostStorageGet()
	

	es := hg.ExternalSettings
	fm := hg.FinancialMetrics
	is := hg.InternalSettings

	// calculate total storage available and remaining
	var totalstorage, storageremaining uint64
	for _, folder := range sg.Folders {
		totalstorage += folder.Capacity
		storageremaining += folder.CapacityRemaining
	}

	// Host Internal Settings
	hostAcceptingContracts.Set(boolToFloat64(is.AcceptingContracts))
	hostTotalStorage.Set(float64(es.TotalStorage))
	hostRemainingStorage.Set(float64(es.RemainingStorage))
	hostWindowSize.Set(float64(is.WindowSize / 6))
	hostCollateralFloat, _ := is.Collateral.Mul(modules.BlockBytesPerMonthTerabyte).Float64()
	hostCollateral.Set(hostCollateralFloat / 1e24)
	hostCollateralBudgetFloat, _ := is.CollateralBudget.Float64()
	hostCollateralBudget.Set(hostCollateralBudgetFloat / 1e24)
	hostMaxCollateralFloat, _ := is.MaxCollateral.Float64()
	hostMaxCollateral.Set(hostMaxCollateralFloat / 1e24)
	hostLockedCollateralFloat, _ := fm.LockedStorageCollateral.Float64()
	hostLockedCollateral.Set(hostLockedCollateralFloat / 1e24)
	hostIngressRevenueFloat, _ := fm.PotentialDownloadBandwidthRevenue.Float64()
	hostIngressRevenue.Set(hostIngressRevenueFloat / 1e24)
	hostEgressRevenueFloat, _ := fm.PotentialUploadBandwidthRevenue.Float64()
	hostEgressRevenue.Set(hostEgressRevenueFloat / 1e24)
	hostStorageRevenueFloat, _ := fm.PotentialStorageRevenue.Float64()
	hostStorageRevenue.Set(hostStorageRevenueFloat / 1e24)
	
	hostContractCount.Set(float64(fm.ContractCount))
	
}

// walletMetrics retrieves and sets the Prometheus metrics related to the
// Sia wallet
func walletMetrics(sc *sia.Client) {
	status, err := sc.WalletGet()
	if errors.Contains(err, ErrAPICallNotRecognized) {
		
		walletModuleLoaded.Set(boolToFloat64(false))
		return
	} else if err != nil {
		
		return
	}
	walletModuleLoaded.Set(boolToFloat64(true))
	if !status.Unlocked {
		walletLocked.Set(boolToFloat64(false))
	}
	walletLocked.Set(boolToFloat64(true))

	ConfirmedBalance, _ := status.ConfirmedSiacoinBalance.Float64()
	walletConfirmedSiacoinBalance.Set(ConfirmedBalance / 1e24)

}