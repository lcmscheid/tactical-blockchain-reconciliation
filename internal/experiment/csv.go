/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

package experiment

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// ResultCSVHeader is the column layout expected by the analysis script and by
// the committed thesis/experiments/results.csv. It must not be reordered
// without also updating the analysis pipeline.
var ResultCSVHeader = []string{
	"strategy",
	"n_forks",
	"partition_duration_s",
	"conflict_level",
	"dep_density",
	"replicate",
	"input_tx_count",
	"unique_id_count",
	"output_tx_count",
	"data_loss_count",
	"data_loss_pct",
	"success_rate",
	"post_consistent",
	"recon_time_ns",
	"recon_time_ms",
	"alloc_delta_b",
	"max_fork_depth",
	"deps_enforced",
	"cycles_detected",
}

// ComplexityCSVHeader is the header for the micro-benchmark output.
var ComplexityCSVHeader = []string{"n", "replicate", "time_ns", "time_us"}

// WriteResults appends every TrialResult to the file at path, emitting the
// header on the first write (creating the file if needed). Subsequent calls
// append without rewriting the header, which keeps the writer safe to use
// across multiple mode invocations of the experiment command.
func WriteResults(path string, rows []TrialResult) error {
	writeHeader := needsHeader(path)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if writeHeader {
		if err := w.Write(ResultCSVHeader); err != nil {
			return err
		}
	}
	for _, r := range rows {
		if err := w.Write(resultRow(r)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// ComplexityRow captures one point of the complexity sweep: the number of
// conflicting transactions `n`, the replicate index within that point, and
// the measured MergeChains latency in both nanoseconds and microseconds.
type ComplexityRow struct {
	N         int
	Replicate int
	TimeNs    int64
	TimeUs    float64
}

// WriteComplexity mirrors WriteResults for the complexity sweep.
func WriteComplexity(path string, rows []ComplexityRow) error {
	writeHeader := needsHeader(path)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if writeHeader {
		if err := w.Write(ComplexityCSVHeader); err != nil {
			return err
		}
	}
	for _, r := range rows {
		row := []string{
			strconv.Itoa(r.N),
			strconv.Itoa(r.Replicate),
			strconv.FormatInt(r.TimeNs, 10),
			strconv.FormatFloat(r.TimeUs, 'f', 3, 64),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func needsHeader(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return info.Size() == 0
}

func resultRow(r TrialResult) []string {
	return []string{
		r.Strategy,
		strconv.Itoa(r.NForks),
		strconv.Itoa(r.PartitionDurationS),
		strconv.FormatFloat(r.ConflictLevel, 'f', 2, 64),
		strconv.FormatFloat(r.DepDensity, 'f', 2, 64),
		strconv.Itoa(r.Replicate),
		strconv.Itoa(r.InputTxCount),
		strconv.Itoa(r.UniqueIDCount),
		strconv.Itoa(r.OutputTxCount),
		strconv.Itoa(r.DataLossCount),
		strconv.FormatFloat(r.DataLossPct, 'f', 6, 64),
		strconv.FormatFloat(r.SuccessRate, 'f', 2, 64),
		fmt.Sprintf("%t", r.PostConsistent),
		strconv.FormatInt(r.ReconTimeNs, 10),
		strconv.FormatFloat(r.ReconTimeMs, 'f', 6, 64),
		strconv.FormatInt(r.AllocDeltaB, 10),
		strconv.Itoa(r.MaxForkDepth),
		strconv.Itoa(r.DepsEnforced),
		strconv.Itoa(r.CyclesDetected),
	}
}
