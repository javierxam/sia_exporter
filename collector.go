package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.sia.tech/siad/modules"
	"go.sia.tech/siad/node/api"
//	"go.sia.tech/siad/types"
	sia "go.sia.tech/siad/node/api/client"
	"gitlab.com/NebulousLabs/errors"
)

var (
	// ErrAPICallNotRecognized is returned by API client calls made to modules that
	// are not yet loaded.
	ErrAPICallNotRecognized = errors.New("API call not recognized")


	// Consensus Metrics
	consensusModuleLoaded = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "consensus_module_loaded", Help: "Is the consensus module loaded. 0=not loaded.  1=loaded"})
	consensusSynced = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "consensus_synced", Help: "Consensus sync status, 0=not synced.  1=synced"})
	consensusHeight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "consensus_height", Help: "Consensus block height"})
	consensusDifficulty = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "consensus_difficulty", Help: "Consensus difficulty"})

	
	// Wallet Metrics
	walletModuleLoaded = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wallet_module_loaded", Help: "Is the wallet module loaded. 0=not loaded.  1=loaded"})
	walletLocked = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wallet_locked", Help: "Is the wallet locked. 0=not locked.  1=locked"})
	walletConfirmedSiacoinBalanceHastings = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wallet_confirmed_siacoin_balance_hastings", Help: "Wallet confirmed Siacoin balance (Hastings)"})
	walletConfirmedSiacoinBalance = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wallet_confirmed_siacoin_balance", Help: "Wallet confirmed Siacoin balance (Siacoins)"})
	walletSiafundBalance = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wallet_siafund_balance", Help: "Wallet Siafund balance"})
	walletSiafundClaimBalance = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wallet_siafund_claim_balance", Help: "Wallet Siafund claim balance"})
	walletNumAddresses = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wallet_num_addresses", Help: "Number of wallet addresses being tracked by Sia"})

	

	// Hostdb Metrics
	hostdbNumAllHosts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hostdb_num_all_hosts", Help: "Total number of hosts in hostdb"})
	hostdbNumActiveHosts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hostdb_num_active_hosts", Help: "Number of active hosts in hostdb"})
	hostdbNumInactiveHosts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hostdb_num_inactive_hosts", Help: "Number of inactive hosts in hostdb"})
	hostdbNumOfflineHosts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hostdb_num_offline_hosts", Help: "Number of offline hosts in hostdb"})

	// Host Metrics
	hostAcceptingContracts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_accepting_contracts", Help: "Is the host accepting contracts 0=no, 1=yes"})
	hostMaxDuration = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_max_duration", Help: "max duration in weeks"})
	hostMaxDownloadBatchSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_max_download_batch_size", Help: "Max Download Batch Size"})
	hostMaxReviseBatchSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "host_max_revise_batch_size", Help: "Max revise Batch Size"})
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
	hg, err := sc.HostGet()
	if errors.Contains(err, ErrAPICallNotRecognized) {
		// Assume module is not loaded if status command is not recognized.
		log.Info("Host module is not loaded")
		return
	} else if err != nil {
		log.Info("Could not fetch host settings")
	}

	sg, err := sc.HostStorageGet()
	if err != nil {
		log.Info("Could not fetch storage info")
	}

	es := hg.ExternalSettings
	fm := hg.FinancialMetrics
	is := hg.InternalSettings
	//	nm := hg.NetworkMetrics

	// calculate total storage available and remaining
	var totalstorage, storageremaining uint64
	for _, folder := range sg.Folders {
		totalstorage += folder.Capacity
		storageremaining += folder.CapacityRemaining
	}

	// convert price from bytes/block to TB/Month
	//	price := is.MinStoragePrice.Mul(modules.BlockBytesPerMonthTerabyte)
	// calculate total revenue
	//	totalRevenue := fm.ContractCompensation.
	//		Add(fm.StorageRevenue).
	//		Add(fm.DownloadBandwidthRevenue).
	//		Add(fm.UploadBandwidthRevenue)
	//	totalPotentialRevenue := fm.PotentialContractCompensation.
	//		Add(fm.PotentialStorageRevenue).
	//		Add(fm.PotentialDownloadBandwidthRevenue).
	//		Add(fm.PotentialUploadBandwidthRevenue)

	// Host Internal Settings
	hostAcceptingContracts.Set(boolToFloat64(is.AcceptingContracts))
	hostTotalStorage.Set(float64(es.TotalStorage))
	hostRemainingStorage.Set(float64(es.RemainingStorage))
	hostMaxDuration.Set(float64(is.MaxDuration))
	hostMaxDownloadBatchSize.Set(float64(is.MaxDownloadBatchSize))
	hostMaxReviseBatchSize.Set(float64(is.MaxReviseBatchSize))
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

// consensuMetrics retrieves and sets the Prometheus metrics related to the
// consensus module
func consensusMetrics(sc *sia.Client) {
	cs, err := sc.ConsensusGet()
	if errors.Contains(err, ErrAPICallNotRecognized) {
		log.Info("Consensus module is not loaded")
		consensusModuleLoaded.Set(boolToFloat64(false))
		return
	} else if err != nil {
		log.Info("Could not get Consensus metrics")
		return
	}

	consensusModuleLoaded.Set(boolToFloat64(true))
	consensusSynced.Set(boolToFloat64(cs.Synced))
	consensusHeight.Set(float64(cs.Height))
	Difficulty, _ := cs.Difficulty.Float64()
	consensusDifficulty.Set(Difficulty)
}

// walletMetrics retrieves and sets the Prometheus metrics related to the
// Sia wallet
func walletMetrics(sc *sia.Client) {
	status, err := sc.WalletGet()
	if errors.Contains(err, ErrAPICallNotRecognized) {
		log.Info("Wallet module is not loaded")
		walletModuleLoaded.Set(boolToFloat64(false))
		return
	} else if err != nil {
		log.Info("Could not get Wallet metrics")
		return
	}
	walletModuleLoaded.Set(boolToFloat64(true))
	if !status.Unlocked {
		walletLocked.Set(boolToFloat64(false))
	}
	walletLocked.Set(boolToFloat64(true))

	ConfirmedBalance, _ := status.ConfirmedSiacoinBalance.Float64()
	walletConfirmedSiacoinBalanceHastings.Set(ConfirmedBalance)
	walletConfirmedSiacoinBalance.Set(ConfirmedBalance / 1e24)

	SiafundBalance, _ := status.SiafundBalance.Float64()
	walletSiafundBalance.Set(SiafundBalance)

	SiafundClaimBalance, _ := status.SiacoinClaimBalance.Float64()
	walletSiafundClaimBalance.Set(SiafundClaimBalance)

	addresses, err := sc.WalletAddressesGet()
	if err != nil {
		log.Info("Could not get wallet addresses")
	}
	walletNumAddresses.Set(float64(len(addresses.Addresses)))
}

// hostdbMetrics retrieves and sets the Prometheus metrics related to the
// Sia hostdb
func hostdbMetrics(sc *sia.Client) {
	hostdb, err := sc.HostDbAllGet()
	if errors.Contains(err, ErrAPICallNotRecognized) {
		log.Info("HostDB module is not loaded")
		return
	} else if err != nil {
		log.Info("Could not get Gateway metrics")
		return
	}

	// Iterate through the hosts and divide by category.
	var activeHosts, inactiveHosts, offlineHosts []api.ExtendedHostDBEntry
	for _, host := range hostdb.Hosts {
		if host.AcceptingContracts && len(host.ScanHistory) > 0 && host.ScanHistory[len(host.ScanHistory)-1].Success {
			activeHosts = append(activeHosts, host)
			continue
		}
		if len(host.ScanHistory) > 0 && host.ScanHistory[len(host.ScanHistory)-1].Success {
			inactiveHosts = append(inactiveHosts, host)
			continue
		}
		offlineHosts = append(offlineHosts, host)
	}

	hostdbNumAllHosts.Set(float64(len(hostdb.Hosts)))
	hostdbNumActiveHosts.Set(float64(len(activeHosts)))
	hostdbNumInactiveHosts.Set(float64(len(inactiveHosts)))
	hostdbNumOfflineHosts.Set(float64(len(offlineHosts)))
}
