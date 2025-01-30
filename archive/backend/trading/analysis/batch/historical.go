package batch

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const (
	ErrDataTooShort = 1001
	ErrPythonExec   = 1002
	ErrJSONParse    = 1003
)

type HistoricalConverter struct {
	PythonPath string
	ScriptPath string
}

func (c *HistoricalConverter) ConvertToBatch(data []float64) (map[string][]float64, error) {
	if len(data) < 26 {
		return map[string][]float64{"error": {ErrDataTooShort}}, nil
	}

	absPath, err := filepath.Abs(c.ScriptPath)
	if err != nil {
		return nil, err
	}

	input, _ := json.Marshal(data)
	// 添加环境变量并设置新的PATH
	cmd := exec.Command(c.PythonPath, absPath, string(input))
	cmd.Env = append(os.Environ(),
		"PYTHONPATH="+filepath.Dir(absPath),
	)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	output, err := cmd.Output()
	if err != nil {
		exitCode := 0
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
		return map[string][]float64{
			"error": {ErrPythonExec, float64(exitCode)},
		}, nil
	}

	// 自定义JSON解析处理NaN值
	var rawResult map[string][]interface{}
	if err := json.Unmarshal(output, &rawResult); err != nil {
		return map[string][]float64{
			"error": {ErrJSONParse},
		}, nil
	}

	finalResult := make(map[string][]float64)
	for key, values := range rawResult {
		var validValues []float64
		for _, v := range values {
			if num, ok := v.(float64); ok && !math.IsNaN(num) {
				validValues = append(validValues, num)
			}
		}
		// 至少保留最后一个有效值
		if len(validValues) > 0 {
			// 过滤前25个初始化值（slow周期26-1）
			startIdx := 25
			if len(validValues) > startIdx {
				finalResult[key] = validValues[startIdx:]
			} else {
				finalResult[key] = validValues
			}
		}
	}
	
	return finalResult, nil
}
