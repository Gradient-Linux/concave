package api

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type cpuCounter struct {
	idle  uint64
	total uint64
}

type cpuSampler struct {
	mu   sync.Mutex
	prev map[string]cpuCounter
}

type memorySnapshot struct {
	Used      uint64 `json:"used"`
	Total     uint64 `json:"total"`
	SwapUsed  uint64 `json:"swap_used"`
	SwapTotal uint64 `json:"swap_total"`
}

var hostCPUSampler = &cpuSampler{prev: map[string]cpuCounter{}}

func cpuMetrics() map[string]any {
	overall, cores, err := hostCPUSampler.sample()
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{
		"overall": overall,
		"cores":   cores,
	}
}

func memoryMetrics() map[string]any {
	snapshot, err := readMemorySnapshot()
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{
		"used":       snapshot.Used,
		"total":      snapshot.Total,
		"swap_used":  snapshot.SwapUsed,
		"swap_total": snapshot.SwapTotal,
	}
}

func (s *cpuSampler) sample() (float64, []map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	current := make(map[string]cpuCounter)
	order := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu") {
			break
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		name := fields[0]
		counter, err := parseCPUCounter(fields[1:])
		if err != nil {
			return 0, nil, err
		}
		current[name] = counter
		order = append(order, name)
	}
	if err := scanner.Err(); err != nil {
		return 0, nil, err
	}

	overall := 0.0
	cores := make([]map[string]any, 0, max(0, len(order)-1))
	for _, name := range order {
		counter := current[name]
		prev, ok := s.prev[name]
		usage := 0.0
		if ok {
			totalDelta := counter.total - prev.total
			idleDelta := counter.idle - prev.idle
			if totalDelta > 0 && idleDelta <= totalDelta {
				usage = (1 - float64(idleDelta)/float64(totalDelta)) * 100
			}
		}
		if name == "cpu" {
			overall = usage
			continue
		}
		cores = append(cores, map[string]any{
			"name":        name,
			"utilization": usage,
		})
	}

	s.prev = current
	return overall, cores, nil
}

func parseCPUCounter(fields []string) (cpuCounter, error) {
	values := make([]uint64, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return cpuCounter{}, fmt.Errorf("parse /proc/stat value %q: %w", field, err)
		}
		values = append(values, value)
	}

	var total uint64
	for _, value := range values {
		total += value
	}

	idle := values[3]
	if len(values) > 4 {
		idle += values[4]
	}
	return cpuCounter{idle: idle, total: total}, nil
}

func readMemorySnapshot() (memorySnapshot, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return memorySnapshot{}, err
	}

	values := make(map[string]uint64)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		values[key] = value * 1024
	}

	total := values["MemTotal"]
	available := values["MemAvailable"]
	if total == 0 {
		return memorySnapshot{}, fmt.Errorf("MemTotal not found in /proc/meminfo")
	}
	if available > total {
		available = total
	}

	swapTotal := values["SwapTotal"]
	swapFree := values["SwapFree"]
	if swapFree > swapTotal {
		swapFree = swapTotal
	}

	return memorySnapshot{
		Used:      total - available,
		Total:     total,
		SwapUsed:  swapTotal - swapFree,
		SwapTotal: swapTotal,
	}, nil
}
