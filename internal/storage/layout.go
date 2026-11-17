package storage

import "fmt"

// LayoutProposal describes a proposed pool layout.
type LayoutProposal struct {
	Layout       string  `json:"layout"`        // single, mirror, raidz1, raidz2
	Drives       int     `json:"drives"`
	UsableBytes  uint64  `json:"usable_bytes"`
	FaultTolerance int   `json:"fault_tolerance"` // number of drive failures survivable
	Explanation  string  `json:"explanation"`
}

// ProposeLayout proposes a layout given drive sizes (bytes).
func ProposeLayout(sizes []uint64) LayoutProposal {
	n := len(sizes)
	if n == 0 {
		return LayoutProposal{Explanation: "no drives selected"}
	}
	var total uint64
	for _, s := range sizes {
		total += s
	}
	// smallest drive determines per-drive unit for raidz; approximate.
	minSize := sizes[0]
	for _, s := range sizes {
		if s < minSize {
			minSize = s
		}
	}

	switch {
	case n == 1:
		return LayoutProposal{
			Layout:          "single",
			Drives:          n,
			UsableBytes:     total,
			FaultTolerance:  0,
			Explanation:     "No redundancy: if this drive fails, data is lost. Consider adding a second drive for a mirror.",
		}
	case n == 2:
		return LayoutProposal{
			Layout:          "mirror",
			Drives:          n,
			UsableBytes:     minSize,
			FaultTolerance:  1,
			Explanation:     "Mirror: keeps a full copy on each drive. Survives 1 drive failure. Usable space equals one drive.",
		}
	case n >= 3:
		expl := fmt.Sprintf("RAIDZ1: survives 1 drive failure. Usable space ≈ %d drives.", n-1)
		if n >= 5 {
			expl += " RAIDZ2 (survives 2 failures) is also offered."
		}
		return LayoutProposal{
			Layout:          "raidz1",
			Drives:          n,
			UsableBytes:     minSize * uint64(n-1),
			FaultTolerance:  1,
			Explanation:     expl,
		}
	default:
		return LayoutProposal{Explanation: "unsupported drive count"}
	}
}
