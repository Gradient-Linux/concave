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

type cpuCoreMetric struct {
	Name        string  `json:"name"`
	Utilization float64 `json:"utilization"`
}

type cpuMetricsPayload struct {
	Overall float64         `json:"overall,omitempty"`
	Cores   []cpuCoreMetric `json:"cores,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type memoryMetricsPayload struct {
	Used      uint64 `json:"used,omitempty"`
	Total     uint64 `json:"total,omitempty"`
	SwapUsed  uint64 `json:"swap_used,omitempty"`
	SwapTotal uint64 `json:"swap_total,omitempty"`
	Error     string `json:"error,omitempty"`
}

type gpuDeviceMetric struct {
	Name        string  `json:"name"`
	Utilization float64 `json:"utilization"`
	MemoryUsed  int64   `json:"memory_used"`
	MemoryTotal int64   `json:"memory_total"`
}

type gpuMetricsPayload struct {
	Devices []gpuDeviceMetric `json:"devices,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type metricsPayload struct {
	Workspace any                  `json:"workspace"`
	Suites    []suiteSummary       `json:"suites"`
	CPU       cpuMetricsPayload    `json:"cpu"`
	GPU       gpuMetricsPayload    `json:"gpu"`
	Memory    memoryMetricsPayload `json:"memory"`
	Timestamp string               `json:"timestamp"`
}

var hostCPUSampler = &cpuSampler{prev: map[string]cpuCounter{}}

func cpuMetrics() cpuMetricsPayload {
	overall, cores, err := hostCPUSampler.sample()
	if err != nil {
		return cpuMetricsPayload{Error: err.Error()}
	}
	return cpuMetricsPayload{
		Overall: overall,
		Cores:   cores,
	}
}

func memoryMetrics() memoryMetricsPayload {
	snapshot, err := readMemorySnapshot()
	if err != nil {
		return memoryMetricsPayload{Error: err.Error()}
	}
	return memoryMetricsPayload{
		Used:      snapshot.Used,
		Total:     snapshot.Total,
		SwapUsed:  snapshot.SwapUsed,
		SwapTotal: snapshot.SwapTotal,
	}
}

func (s *cpuSampler) sample() (float64, []cpuCoreMetric, error) {
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
	cores := make([]cpuCoreMetric, 0, max(0, len(order)-1))
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
		cores = append(cores, cpuCoreMetric{Name: name, Utilization: usage})
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
