package antientropy

// Compares peer and local NodeSummaries
// Compiles a list of SyncJobs that should be run to sync
func DiffSummaries(local, remote NodeSummary, peerID string) []SyncJob {

	// New list to hold sync jobs
	var jobs []SyncJob

	// Iterate over remote buckets
	for bucket, peerSumm := range remote {

		// Check if we have this bucket locally
		localSumm, ok := local[bucket]
		if !ok {
			// No such bucket locally; sync all bins
			for _, bin := range peerSumm.Bins {
				jobs = append(jobs, SyncJob{PeerID: peerID, Bucket: bucket, Bin: bin.Bin})
			}
			continue
		}

		// We have the bucket locally, compare bins
		localBins := make(map[BinID]*BinSummary)
		for _, bin := range localSumm.Bins {
			localBins[bin.Bin] = &bin
		}

		// Iterate over peer bins and check if we need to sync
		// NOTE that we only check the count of points in the bin!!
		// This may be an issue later...
		for _, peerBin := range peerSumm.Bins {
			localBin, exists := localBins[peerBin.Bin]
			if !exists || localBin.Count < peerBin.Count {
				jobs = append(jobs, SyncJob{PeerID: peerID, Bucket: bucket, Bin: peerBin.Bin})
				continue
			}
		}
	}
	return jobs
}
